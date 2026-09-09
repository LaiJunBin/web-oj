package handler

import (
	"bytes"
	"io"
	"testing"
	"time"

	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	logreader "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/log_reader"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	event "judge-backend/internal/app/evaluation_environment_domain/events"
	dockermocks "judge-backend/internal/mocks/github.com/docker/docker/client"
	"judge-backend/internal/projectpath"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// White-box (`package handler`) for the same reason as the readerMap test:
// these scenarios need deterministic sequencing around unexported state
// (buildImage is the unexported bus callback, and the interleavings under
// test cannot be ordered reliably from outside the package).

type ownershipFixture struct {
	handler    *BuildEventHandler
	imageRepo  *mocks.ImageRepository
	buildRepo  *mocks.BuildRepository
	dockerMock *dockermocks.MockAPIClient
	fss        *sharedmocks.FileSystemService
	build      *build.Build
	image      *image.Image
	event      *event.BuildImageEvent
}

func newOwnershipFixture(t *testing.T, tag string) *ownershipFixture {
	t.Helper()
	f := &ownershipFixture{
		imageRepo:  mocks.NewImageRepository(t),
		buildRepo:  mocks.NewBuildRepository(t),
		dockerMock: dockermocks.NewMockAPIClient(t),
		fss:        sharedmocks.NewFileSystemService(t),
	}
	f.handler = &BuildEventHandler{
		imageRepo:    f.imageRepo,
		buildRepo:    f.buildRepo,
		client:       f.dockerMock,
		fss:          f.fss,
		logger:       sharedlogger.NewMockLogger(),
		readerMap:    map[string]*logreader.LogReader{},
		activeBuilds: map[string]struct{}{},
	}

	f.image = image.MustNew(&image.Config{
		Tag:         tag,
		Description: "ownership test image",
		StoragePath: projectpath.Root + "/build_service_test",
		MountPath:   "/var/www/html",
	})
	f.build = build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew(tag),
		Status: buildstatus.PENDING,
		Batch:  batch.MustNew(1),
	})
	f.event = event.NewBuildImageEvent(&event.BuildImageEventConfig{
		Tag:     f.build.Tag(),
		Batch:   f.build.Batch().Value(),
		Path:    f.image.Project().StoragePath(),
		Options: &event.BuildImageOptions{},
	})

	// FindByTag / FindByTagAndBatch hand every invocation the SAME entity —
	// exactly the production aliasing the ownership boundary exists for.
	f.imageRepo.On("FindByTag", tag).Return(f.image, nil)
	f.buildRepo.On("FindByTagAndBatch", tag, f.build.Batch().Value()).Return(f.build, nil)
	f.fss.On("Exists", f.image.Project().StoragePath()).Return(true)
	f.fss.On("TarPath", f.image.Project().StoragePath()).Return(&bytes.Buffer{}, nil)
	return f
}

// TestBuildImage_DuplicateEventsExecuteExactlyOnce pins the execution-
// ownership boundary: the build queue has multiple workers, the bus spawns
// one goroutine per event, and the repository returns the same entity to all
// of them. Two events for one tag+batch must produce exactly ONE docker
// build and exactly ONE commit — the strict .Once() expectations below make
// a second ImageBuild or Commit call fail the test. (The old code let both
// invocations pass the status check — the second's rejected SetStatus was
// ignored — so both executed; and even a loser stopped by the status CAS
// alone would still run its deferred Commit, deleting the in-flight build
// from the repository's buildMap and making it uncancellable.)
func TestBuildImage_DuplicateEventsExecuteExactlyOnce(t *testing.T) {
	f := newOwnershipFixture(t, "ownership-dup:latest")

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	f.dockerMock.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: pipeReader}, nil).Once()
	f.buildRepo.On("Commit", f.build).Return(nil).Once()
	f.imageRepo.On("Save", f.image).Return(nil).Once()

	ownerDone := make(chan struct{})
	go func() {
		defer close(ownerDone)
		f.handler.buildImage(f.event)
	}()

	// Wait until the owner has genuinely claimed and started executing.
	require.Eventually(t, func() bool {
		f.handler.readerMu.Lock()
		defer f.handler.readerMu.Unlock()
		_, active := f.handler.activeBuilds[f.handler.buildKey(f.build)]
		return active
	}, 5*time.Second, 5*time.Millisecond, "owner never claimed the build")

	// The duplicate event arrives while the owner is executing: it must
	// return without a second ImageBuild and without a second Commit.
	f.handler.buildImage(f.event)

	pipeWriter.Close()
	select {
	case <-ownerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("owner build did not finish after its log stream closed")
	}

	assert.True(t, f.build.IsSuccess(), "the owner's outcome must stand, got %s", f.build.Status().String())
}

// TestBuildImage_StaleEventCannotExecuteAnotherBuild pins the job identity
// the event now carries: build A (batch 1) was cancelled while its event
// still sat in the queue, and build B (batch 2) is the tag's current pending
// build. A's stale event must resolve ITS OWN build via FindByTagAndBatch —
// find it terminal, stand down, commit nothing — never pick up B; B's own
// event then executes B exactly once. (The old handler resolved events with
// FindLastBuild(tag), so A's stale event would have grabbed and run B.)
func TestBuildImage_StaleEventCannotExecuteAnotherBuild(t *testing.T) {
	tag := "ownership-stale:latest"
	imageRepo := mocks.NewImageRepository(t)
	buildRepo := mocks.NewBuildRepository(t)
	dockerMock := dockermocks.NewMockAPIClient(t)
	fss := sharedmocks.NewFileSystemService(t)

	h := &BuildEventHandler{
		imageRepo:    imageRepo,
		buildRepo:    buildRepo,
		client:       dockerMock,
		fss:          fss,
		logger:       sharedlogger.NewMockLogger(),
		readerMap:    map[string]*logreader.LogReader{},
		activeBuilds: map[string]struct{}{},
	}

	img := image.MustNew(&image.Config{
		Tag:         tag,
		Description: "stale event test image",
		StoragePath: projectpath.Root + "/build_service_test",
		MountPath:   "/var/www/html",
	})
	cancelledA := build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew(tag),
		Status: buildstatus.CANCELLED,
		Batch:  batch.MustNew(1),
	})
	pendingB := build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew(tag),
		Status: buildstatus.PENDING,
		Batch:  batch.MustNew(2),
	})

	imageRepo.On("FindByTag", tag).Return(img, nil)
	buildRepo.On("FindByTagAndBatch", tag, 1).Return(cancelledA, nil).Once()
	buildRepo.On("FindByTagAndBatch", tag, 2).Return(pendingB, nil).Once()

	eventFor := func(b *build.Build) *event.BuildImageEvent {
		return event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:     b.Tag(),
			Batch:   b.Batch().Value(),
			Path:    img.Project().StoragePath(),
			Options: &event.BuildImageOptions{},
		})
	}

	// A's stale event first, alone: it must execute nothing and commit
	// nothing — no ImageBuild/Commit/Exists expectations have been consumed
	// yet, and B must still be pending afterwards.
	h.buildImage(eventFor(cancelledA))
	require.True(t, pendingB.IsPending(), "the stale event must not have executed build B")
	require.True(t, cancelledA.IsCancelled(), "the stale event must not have resurrected its own cancelled build")

	// B's own event executes B, exactly once.
	fss.On("Exists", img.Project().StoragePath()).Return(true).Once()
	fss.On("TarPath", img.Project().StoragePath()).Return(&bytes.Buffer{}, nil).Once()
	dockerMock.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: io.NopCloser(bytes.NewReader(nil))}, nil).Once()
	buildRepo.On("Commit", pendingB).Return(nil).Once()
	imageRepo.On("Save", img).Return(nil).Once()

	h.buildImage(eventFor(pendingB))

	assert.True(t, pendingB.IsSuccess(), "B's own event must have executed it, got %s", pendingB.Status().String())
}

// TestBuildImage_CancelDuringExecution_CancelWinsAndOwnerDoesNotOverride
// drives the full cancel path against an EXECUTING build — the scenario the
// old same-key readerMap test never covered (it only protected the second
// build's reader): cancelBuild atomically moves the shared entity to
// CANCELLED, closes the owner's reader via the readerMap, and the owner —
// waking up from execute — must NOT override the decided status with
// SUCCESS. (The entity-level memory-safety under truly overlapping access
// is pinned separately by TestBuild_SetStatus_AtomicCompareAndSet — this
// test's steps are sequenced too deterministically to overlap.)
func TestBuildImage_CancelDuringExecution_CancelWinsAndOwnerDoesNotOverride(t *testing.T) {
	f := newOwnershipFixture(t, "ownership-cancel:latest")

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	f.dockerMock.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: pipeReader}, nil).Once()
	// Two commits: the cancel handler persists CANCELLED, then the owner's
	// deferred commit persists the final state (still CANCELLED).
	f.buildRepo.On("Commit", f.build).Return(nil).Twice()
	// No imageRepo.Save expectation: the build ends neither SUCCESS nor
	// ERROR, so an image state write would panic the strict mock.

	ownerDone := make(chan struct{})
	go func() {
		defer close(ownerDone)
		f.handler.buildImage(f.event)
	}()

	// Wait until the owner has registered its reader — from that point the
	// cancel path has something to close.
	require.Eventually(t, func() bool {
		f.handler.readerMu.Lock()
		defer f.handler.readerMu.Unlock()
		return f.handler.readerMap[f.handler.buildKey(f.build)] != nil
	}, 5*time.Second, 5*time.Millisecond, "owner never registered its log reader")

	f.handler.cancelBuild(event.NewCancelBuildEvent(f.build))

	select {
	case <-ownerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("cancelling did not unblock the executing build")
	}

	assert.True(t, f.build.IsCancelled(), "cancel won the transition; the owner must not override it, got %s", f.build.Status().String())
}
