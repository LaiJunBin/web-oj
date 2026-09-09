package handler_test

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/events"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/handler"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"judge-backend/internal/mocks/github.com/docker/docker/client"
	"judge-backend/internal/projectpath"
	"sync"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type buildHandlerMocks struct {
	eventbus  *sharedmocks.EventBus
	fss       *sharedmocks.FileSystemService
	buildRepo *mocks.BuildRepository
	imageRepo *mocks.ImageRepository
	client    *client.MockAPIClient
}

func newBuildHandlerMocks(t *testing.T) *buildHandlerMocks {
	return &buildHandlerMocks{
		eventbus:  sharedmocks.NewEventBus(t),
		fss:       sharedmocks.NewFileSystemService(t),
		buildRepo: mocks.NewBuildRepository(t),
		imageRepo: mocks.NewImageRepository(t),
		client:    client.NewMockAPIClient(t),
	}
}

func setupBuildHandlerMocks(t *testing.T) *buildHandlerMocks {
	mocks := newBuildHandlerMocks(t)
	return mocks
}

func registerBuildEventHandler(mocks *buildHandlerMocks) {
	handler.RegisterBuildEventHandler(mocks.eventbus, mocks.imageRepo, mocks.buildRepo, mocks.client, mocks.fss, logger.NewMockLogger())
}

func createMockTarBuffer() *bytes.Buffer {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"Dockerfile", "FROM nginx:latest\n"},
		{"index.html", "Hello, World!"},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}

		err := tw.WriteHeader(hdr)
		if err != nil {
			panic(err)
		}

		_, err = tw.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}

	err := tw.Close()
	if err != nil {
		panic(err)
	}

	return buf
}

func TestSubscribeBuildImage(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		mocks := setupBuildHandlerMocks(t)
		mockImage := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})
		mockBuild := build.New(&build.BuildConfig{
			Tag:    mockImage.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(1),
		})
		mockEvent := event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:   mockBuild.Tag(),
			Batch: mockBuild.Batch().Value(),
			Path:  mockImage.Project().StoragePath(),
			Options: &event.BuildImageOptions{
				NoCache: true,
			},
		})
		ctx := context.Background()
		tarBuffer := createMockTarBuffer()
		options := types.ImageBuildOptions{
			Tags:    []string{mockImage.Tag().String()},
			NoCache: true,
		}
		mocks.imageRepo.On("FindByTag", mockImage.Tag().String()).Return(mockImage, nil)
		mocks.buildRepo.On("FindByTagAndBatch", mockImage.Tag().String(), mockBuild.Batch().Value()).Return(mockBuild, nil)
		mocks.buildRepo.On("Commit", mockBuild).Return(nil)
		mocks.imageRepo.On("Save", mockImage).Return(nil)
		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(true)
		mocks.fss.On("TarPath", mockImage.Project().StoragePath()).Return(tarBuffer, nil)
		mocks.client.On("ImageBuild", ctx, tarBuffer, options).Return(types.ImageBuildResponse{
			Body: io.NopCloser(bytes.NewReader([]byte(`
					{"stream":"Step 1/5 : FROM alpine:3.12\n"}
				`))),
		}, nil)

		mocks.eventbus.On("SubscribeAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			handler, ok := args.Get(0).(func(event *event.BuildImageEvent))
			if ok {
				handler(mockEvent)

				assert.Equal(t, buildstatus.SUCCESS, mockBuild.Status())
				assert.True(t, mockImage.IsReady())
				assert.Equal(t, len(mockBuild.Logs()), 1)
			}
		}).Return(nil)

		registerBuildEventHandler(mocks)
	})

	t.Run("Fail if image project not found", func(t *testing.T) {
		mocks := setupBuildHandlerMocks(t)
		mockImage := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})
		mockBuild := build.New(&build.BuildConfig{
			Tag:    mockImage.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(1),
		})
		mockEvent := event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:     mockBuild.Tag(),
			Batch:   mockBuild.Batch().Value(),
			Path:    mockImage.Project().StoragePath(),
			Options: &event.BuildImageOptions{},
		})
		mocks.imageRepo.On("FindByTag", mockImage.Tag().String()).Return(mockImage, nil)
		mocks.buildRepo.On("FindByTagAndBatch", mockImage.Tag().String(), mockBuild.Batch().Value()).Return(mockBuild, nil)
		mocks.buildRepo.On("Commit", mockBuild).Return(nil)
		mocks.imageRepo.On("Save", mockImage).Return(nil)
		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(false)
		mocks.eventbus.On("SubscribeAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			handler, ok := args.Get(0).(func(event *event.BuildImageEvent))
			if ok {
				handler(mockEvent)

				assert.Equal(t, buildstatus.ERROR, mockBuild.Status())
				assert.False(t, mockImage.IsReady())
			}
		}).Return(nil)
		registerBuildEventHandler(mocks)
	})

	t.Run("Fail if image build failed", func(t *testing.T) {
		mocks := setupBuildHandlerMocks(t)
		mockImage := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})
		mockBuild := build.New(&build.BuildConfig{
			Tag:    mockImage.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(1),
		})
		mockEvent := event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:   mockBuild.Tag(),
			Batch: mockBuild.Batch().Value(),
			Path:  mockImage.Project().StoragePath(),
			Options: &event.BuildImageOptions{
				NoCache: true,
			},
		})
		ctx := context.Background()
		tarBuffer := createMockTarBuffer()
		options := types.ImageBuildOptions{
			Tags:    []string{mockImage.Tag().String()},
			NoCache: true,
		}
		mocks.imageRepo.On("FindByTag", mockImage.Tag().String()).Return(mockImage, nil)
		mocks.buildRepo.On("FindByTagAndBatch", mockImage.Tag().String(), mockBuild.Batch().Value()).Return(mockBuild, nil)
		mocks.buildRepo.On("Commit", mockBuild).Return(nil)
		mocks.imageRepo.On("Save", mockImage).Return(nil)
		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(true)
		mocks.fss.On("TarPath", mockImage.Project().StoragePath()).Return(tarBuffer, nil)
		mocks.client.On("ImageBuild", ctx, tarBuffer, options).Return(types.ImageBuildResponse{}, fmt.Errorf("failed to build image"))

		mocks.eventbus.On("SubscribeAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			handler, ok := args.Get(0).(func(event *event.BuildImageEvent))
			if ok {
				handler(mockEvent)

				assert.Equal(t, buildstatus.ERROR, mockBuild.Status())
				assert.False(t, mockImage.IsReady())
			}
		}).Return(nil)

		registerBuildEventHandler(mocks)
	})

	// The log stream is the ONLY evidence of a build's outcome — success is
	// inferred from the absence of Error-level entries in it. So a stream
	// that breaks midway leaves no verdict at all, and answering SUCCESS
	// would hand the admin an image whose build was never observed.
	t.Run("Fail if the build log stream breaks midway", func(t *testing.T) {
		mocks := setupBuildHandlerMocks(t)
		mockImage := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})
		mockBuild := build.New(&build.BuildConfig{
			Tag:    mockImage.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(1),
		})
		mockEvent := event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:   mockBuild.Tag(),
			Batch: mockBuild.Batch().Value(),
			Path:  mockImage.Project().StoragePath(),
			Options: &event.BuildImageOptions{
				NoCache: true,
			},
		})
		ctx := context.Background()
		tarBuffer := createMockTarBuffer()
		options := types.ImageBuildOptions{
			Tags:    []string{mockImage.Tag().String()},
			NoCache: true,
		}
		mocks.imageRepo.On("FindByTag", mockImage.Tag().String()).Return(mockImage, nil)
		mocks.buildRepo.On("FindByTagAndBatch", mockImage.Tag().String(), mockBuild.Batch().Value()).Return(mockBuild, nil)
		mocks.buildRepo.On("Commit", mockBuild).Return(nil)
		mocks.imageRepo.On("Save", mockImage).Return(nil)
		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(true)
		mocks.fss.On("TarPath", mockImage.Project().StoragePath()).Return(tarBuffer, nil)
		// One complete line, then the daemon connection dies.
		mocks.client.On("ImageBuild", ctx, tarBuffer, options).Return(types.ImageBuildResponse{
			Body: &brokenStreamBody{
				data: []byte(`{"stream":"Step 1/5 : FROM alpine:3.12\n"}` + "\n"),
				err:  errors.New("connection reset by peer"),
			},
		}, nil)

		mocks.eventbus.On("SubscribeAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			handler, ok := args.Get(0).(func(event *event.BuildImageEvent))
			if ok {
				handler(mockEvent)

				assert.Equal(t, buildstatus.ERROR, mockBuild.Status())
				assert.False(t, mockImage.IsReady())
			}
		}).Return(nil)

		registerBuildEventHandler(mocks)
	})
}

// brokenStreamBody serves data and then fails, modelling a docker daemon
// connection that dies part-way through a build's log stream.
type brokenStreamBody struct {
	data []byte
	off  int
	err  error
}

func (b *brokenStreamBody) Read(p []byte) (int, error) {
	if b.off < len(b.data) {
		n := copy(p, b.data[b.off:])
		b.off += n
		return n, nil
	}
	return 0, b.err
}

func (b *brokenStreamBody) Close() error { return nil }

// TestReaderMapConcurrentAccess is a regression test for a fatal
// "concurrent map read/write" crash: buildImage() writes to readerMap while
// cancelBuild() reads/deletes from it, and the two run on separate
// SubscribeAsync goroutines with no synchronization. It fires many
// build/cancel pairs targeting the SAME map key concurrently under `go test
// -race`; without the readerMu mutex this reliably trips the race detector
// (or, outside of -race, can crash the whole process with a runtime fatal
// error that no recover() can catch).
func TestReaderMapConcurrentAccess(t *testing.T) {
	mocks := setupBuildHandlerMocks(t)

	var buildImageHandler func(event *event.BuildImageEvent)
	var cancelBuildHandler func(event *event.CancelBuildEvent)
	mocks.eventbus.On("SubscribeAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if h, ok := args.Get(0).(func(event *event.BuildImageEvent)); ok {
			buildImageHandler = h
		}
		if h, ok := args.Get(0).(func(event *event.CancelBuildEvent)); ok {
			cancelBuildHandler = h
		}
	}).Return(nil)

	registerBuildEventHandler(mocks)
	require.NotNil(t, buildImageHandler)
	require.NotNil(t, cancelBuildHandler)

	tarBuffer := createMockTarBuffer()
	logsBody := []byte("{\"stream\":\"step 1\\n\"}\n{\"stream\":\"step 2\\n\"}\n")

	// Commit/Save take a pointer to a live *Build/*Image that its owning
	// goroutine keeps mutating throughout the run. Matching by mock.Anything
	// (rather than a specific pointer per iteration) keeps testify's
	// call-matching from reflecting into those objects while they are
	// concurrently written — that would be a race in the TEST's own
	// bookkeeping, not the production code this test targets.
	mocks.buildRepo.On("Commit", mock.Anything).Return(nil)
	mocks.imageRepo.On("Save", mock.Anything).Return(nil).Maybe()

	const iterations = 50
	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		tag := fmt.Sprintf("race-build-%d:latest", i)
		mockImage := image.MustNew(&image.Config{
			Tag:         tag,
			Description: "description",
			StoragePath: projectpath.Root + "/build_service_test",
			MountPath:   "/var/www/html",
		})
		mockBuild := build.New(&build.BuildConfig{
			Tag:    mockImage.Tag(),
			Status: buildstatus.PENDING,
			Batch:  batch.MustNew(1),
		})
		mockEvent := event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:   mockBuild.Tag(),
			Batch: mockBuild.Batch().Value(),
			Path:  mockImage.Project().StoragePath(),
			Options: &event.BuildImageOptions{
				NoCache: true,
			},
		})
		options := types.ImageBuildOptions{Tags: []string{tag}, NoCache: true}

		mocks.imageRepo.On("FindByTag", tag).Return(mockImage, nil)
		mocks.buildRepo.On("FindByTagAndBatch", tag, mockBuild.Batch().Value()).Return(mockBuild, nil)
		mocks.fss.On("Exists", mockImage.Project().StoragePath()).Return(true)
		mocks.fss.On("TarPath", mockImage.Project().StoragePath()).Return(tarBuffer, nil)
		mocks.client.On("ImageBuild", context.Background(), tarBuffer, options).Return(types.ImageBuildResponse{
			Body: io.NopCloser(bytes.NewReader(logsBody)),
		}, nil)

		// A dedicated build entity (same Tag+Batch, hence the same readerMap
		// key, but a distinct object — no entity-level state is shared with
		// mockBuild) so cancelBuild races the SAME map key that buildImage
		// is writing, without racing the *build.Build entity itself.
		cancelTarget := build.New(&build.BuildConfig{
			Tag:    mockImage.Tag(),
			Status: buildstatus.BUILDING,
			Batch:  batch.MustNew(1),
		})
		cancelEvent := event.NewCancelBuildEvent(cancelTarget)

		wg.Add(2)
		go func() {
			defer wg.Done()
			buildImageHandler(mockEvent)
		}()
		go func() {
			defer wg.Done()
			cancelBuildHandler(cancelEvent)
		}()
	}
	wg.Wait()
}
