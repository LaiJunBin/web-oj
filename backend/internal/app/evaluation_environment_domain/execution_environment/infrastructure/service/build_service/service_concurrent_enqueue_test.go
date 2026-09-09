package buildservice_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	domainrepo "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	domainservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	infrarepo "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/build_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/projectpath"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestBuildService_SetupFailureReleasesReservation_NextEnqueueSucceeds is
// the repository half of decisive test 3: a reservation whose setup failed
// and was released must leave the tag reservable — without the release the
// tag would be locked forever and every later enqueue would fail. Uses the
// REAL repository so the release-then-reserve cycle exercises the actual
// identity-guarded map, and also pins the guard itself: releasing the OLD
// reservation after a new one was taken must not evict the new one.
func TestBuildService_SetupFailureReleasesReservation_NextEnqueueSucceeds(t *testing.T) {
	tag := "release-retry:latest"
	img := image.MustNew(&image.Config{
		Tag:         tag,
		Description: "release retry test",
		StoragePath: projectpath.Root + "/build_service_test",
		MountPath:   "/var/www/html",
	})

	repo := infrarepo.NewBuildRepository(nil)
	factory := mocks.NewBuildFactory(t)
	newPending := func(batchValue int) *build.Build {
		return build.New(&build.BuildConfig{
			Tag:    img.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(batchValue),
		})
	}
	first := newPending(1)
	discarded := newPending(1) // the rejected attempt still allocates before the atomic check refuses it
	second := newPending(1)
	factory.On("Create", img).Return(first, nil).Once()
	factory.On("Create", img).Return(discarded, nil).Once()
	factory.On("Create", img).Return(second, nil).Once()

	fss := sharedmocks.NewFileSystemService(t)
	queue := mocks.NewBuildQueue(t)
	svc := buildservice.NewService(sharedmocks.NewEventBus(t), queue, fss, factory, repo).(interface {
		ReserveBuild(*image.Image) (*build.Build, error)
		ReleaseReservedBuild(*build.Build)
	})

	reserved, err := svc.ReserveBuild(img)
	require.NoError(t, err)
	require.Equal(t, first, reserved)

	// While the reservation is held, a second reserve must fail.
	_, err = svc.ReserveBuild(img)
	require.ErrorIs(t, err, domainrepo.ErrorBuildAlreadyInProgress)

	// Setup failed → release. The next reservation must succeed.
	svc.ReleaseReservedBuild(reserved)
	retried, err := svc.ReserveBuild(img)
	require.NoError(t, err, "after a release the tag must be reservable again")
	require.Equal(t, second, retried)

	// The identity guard: releasing the OLD, already-released reservation
	// again must not evict the new one.
	svc.ReleaseReservedBuild(reserved)
	assert.Equal(t, second, repo.FindCurrentBuildByTag(tag), "a stale release must not evict the current reservation")
}

// TestBuildService_ConcurrentEnqueueSameTag_OnlyOnePendingSurvives closes
// the concurrent-rebuild TOCTOU end to end: two racing enqueues for the same
// tag both reach the factory (which, reading only committed builds,
// allocates them the SAME batch number — the mocked factory reproduces
// exactly that), and under the old plain Save the second registration
// silently overwrote the first's pending build, leaving two enqueued events
// pointing at one map slot. With SavePendingExclusive the check and the
// registration are one atomic repository step: exactly one enqueue may
// succeed, the loser gets ErrorBuildAlreadyInProgress, and the surviving
// pending build is the winner's.
func TestBuildService_ConcurrentEnqueueSameTag_OnlyOnePendingSurvives(t *testing.T) {
	tag := "concurrent-enqueue:latest"
	img := image.MustNew(&image.Config{
		Tag:         tag,
		Description: "concurrent enqueue test",
		StoragePath: projectpath.Root + "/build_service_test",
		MountPath:   "/var/www/html",
	})

	// The real repository (map paths only — SavePendingExclusive never
	// touches SQL) is the component under test; mocking it would only
	// re-assert the test's own assumptions about its atomicity.
	repo := infrarepo.NewBuildRepository(nil)

	factory := mocks.NewBuildFactory(t)
	newPending := func() *build.Build {
		return build.New(&build.BuildConfig{
			Tag:    img.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(1),
		})
	}
	factory.On("Create", img).Return(newPending(), nil).Once()
	factory.On("Create", img).Return(newPending(), nil).Once()

	fss := sharedmocks.NewFileSystemService(t)
	fss.On("Exists", img.Project().StoragePath()).Return(true)

	var enqueued atomic.Int32
	queue := mocks.NewBuildQueue(t)
	queue.On("Enqueue", mock.Anything).Run(func(mock.Arguments) {
		enqueued.Add(1)
	}).Maybe()

	svc := buildservice.NewService(sharedmocks.NewEventBus(t), queue, fss, factory, repo)

	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			errs[idx] = svc.EnqueueBuildJob(img, &domainservice.BuildOptions{})
		}(i)
	}
	close(start)
	wg.Wait()

	winners, losers := 0, 0
	for _, err := range errs {
		switch err {
		case nil:
			winners++
		case domainrepo.ErrorBuildAlreadyInProgress:
			losers++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	require.Equal(t, 1, winners, "exactly one concurrent enqueue may register its pending build")
	require.Equal(t, 1, losers, "the other must be rejected, not silently dropped or doubled")
	assert.EqualValues(t, 1, enqueued.Load(), "exactly one build event may reach the queue")
	assert.NotNil(t, repo.FindCurrentBuildByTag(tag), "the winner's pending build must survive in the repository")
}
