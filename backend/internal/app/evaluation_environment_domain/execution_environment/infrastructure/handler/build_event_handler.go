package handler

import (
	"context"
	"errors"
	"fmt"
	sharedservice "judge-backend/internal/app/shared/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/events"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	logreader "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/log_reader"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type BuildEventHandler struct {
	imageRepo repository.ImageRepository
	buildRepo repository.BuildRepository
	client    client.APIClient
	fss       sharedservice.FileSystemService
	logger    sharedtypes.Logger
	// readerMu guards readerMap and activeBuilds: buildImage and cancelBuild
	// run on separate async event-bus goroutines, one per published event.
	readerMu  sync.Mutex
	readerMap map[string]*logreader.LogReader
	// activeBuilds is the execution-ownership boundary: at most one
	// buildImage invocation may execute a given tag+batch at a time. The
	// build queue has multiple workers and the bus spawns a goroutine per
	// event, so duplicate/concurrent events for the same build are a normal
	// occurrence, and FindLastBuild hands every caller the SAME entity —
	// without this claim, both invocations passed the status check and both
	// ran the docker build.
	activeBuilds map[string]struct{}
}

func RegisterBuildEventHandler(
	eventBus sharedtypes.EventBus,
	imageRepo repository.ImageRepository,
	buildRepo repository.BuildRepository,
	client client.APIClient,
	fss sharedservice.FileSystemService,
	logger sharedtypes.Logger,
) sharedtypes.EventHandler {
	handler := &BuildEventHandler{
		imageRepo:    imageRepo,
		buildRepo:    buildRepo,
		client:       client,
		fss:          fss,
		logger:       logger,
		readerMap:    map[string]*logreader.LogReader{},
		activeBuilds: map[string]struct{}{},
	}

	eventBus.SubscribeAsync(handler.buildImage, false)
	eventBus.SubscribeAsync(handler.cancelBuild, false)

	return handler
}

func (b *BuildEventHandler) buildKey(build *build.Build) string {
	return fmt.Sprintf("%s-%d", build.Tag(), build.Batch().Value())
}

func (b *BuildEventHandler) buildImage(event *event.BuildImageEvent) {
	image, err := b.imageRepo.FindByTag(event.Tag())
	if err != nil {
		return
	}

	// FindByTagAndBatch, not FindLastBuild: the event carries the identity
	// of the ONE build it was enqueued for. By consumption time the "last"
	// build of this tag can be a different, newer build — a stale event must
	// find its own (already terminal, or gone) build and stand down, never
	// execute someone else's.
	build, err := b.buildRepo.FindByTagAndBatch(event.Tag(), event.Batch())
	if err != nil {
		return
	}

	if build == nil {
		return
	}

	// Claim execution ownership of this tag+batch before touching the build.
	// A second event for the same build (duplicate publish, or a concurrent
	// worker) returns here instead of racing the owner on the SAME entity.
	key := b.buildKey(build)
	b.readerMu.Lock()
	if _, active := b.activeBuilds[key]; active {
		b.readerMu.Unlock()
		// Duplicate/concurrent events for the same build are a normal
		// occurrence (see activeBuilds above) — not worth more than Debug.
		b.logger.Debugf("[BuildEventHandler] skipping duplicate build execution for %s", key)
		return
	}
	b.activeBuilds[key] = struct{}{}
	b.readerMu.Unlock()
	defer func() {
		b.readerMu.Lock()
		delete(b.activeBuilds, key)
		b.readerMu.Unlock()
	}()

	// SetStatus is an atomic compare-and-set on the entity; a rejected
	// BUILDING transition (already cancelled, already terminal) means this
	// invocation does not own the execution and must not run it. The old
	// code checked CanChangeTo separately and then ignored SetStatus's error
	// — under concurrency both callers passed the check, the second
	// SetStatus failed silently, and BOTH executed. This runs BEFORE the
	// commit defer below is registered, so a stood-down invocation commits
	// nothing and deletes nothing.
	if serr := build.SetStatus(buildstatus.BUILDING); serr != nil {
		b.logger.Infof("[BuildEventHandler] skipping build execution for %s - not in a buildable state: %v", key, serr)
		return
	}

	defer func() {
		if err != nil {
			// A rejected transition (e.g. the cancel handler already moved
			// the build to CANCELLED) means the ERROR outcome is not ours to
			// record — leave the status the other writer decided.
			if serr := build.SetStatus(buildstatus.ERROR); serr != nil {
				b.logger.Warnf("[BuildEventHandler] skipping ERROR status for build %s - status already decided: %v", key, serr)
			}
		}

		b.buildRepo.Commit(build)
		if build.IsSuccess() {
			image.MarkAsReady()
			b.imageRepo.Save(image)
		} else if build.IsError() {
			image.MarkAsNotReady()
			b.imageRepo.Save(image)
		}
	}()

	if err = b.execute(build, event.Path(), &buildservice.BuildOptions{
		NoCache: event.Options().NoCache,
	}); err != nil {
		return
	}

	// Rejected here means the build was cancelled while executing — the
	// CANCELLED status stands, and the deferred commit persists it.
	if serr := build.SetStatus(buildstatus.SUCCESS); serr != nil {
		b.logger.Infof("[BuildEventHandler] skipping SUCCESS status for build %s - status already decided: %v", key, serr)
	}
}

func (b *BuildEventHandler) execute(build *build.Build, projectPath string, options *buildservice.BuildOptions) error {
	if !b.fss.Exists(projectPath) {
		return buildservice.ErrorImageProjectNotSet
	}

	tar, err := b.fss.TarPath(projectPath)
	if err != nil {
		return err
	}

	response, err := b.client.ImageBuild(context.Background(), tar, types.ImageBuildOptions{
		Tags:    []string{build.Tag()},
		NoCache: options.NoCache,
	})
	if err != nil {
		build.AddLog(buildlog.New(buildloglevel.Error, err.Error(), time.Now()))
		b.logger.Errorf("[BuildEventHandler] error building image %s: %v", build.Tag(), err)
		return err
	}

	reader := logreader.New(response.Body).(*logreader.LogReader)
	key := b.buildKey(build)
	b.readerMu.Lock()
	b.readerMap[key] = reader
	b.readerMu.Unlock()
	defer func() {
		reader.Close()
		b.readerMu.Lock()
		// Delete only if the key still points at OUR reader: a concurrent
		// build of the same tag overwrites the entry, and unconditionally
		// deleting here would remove that later build's reader — leaving it
		// impossible to cancel.
		if b.readerMap[key] == reader {
			delete(b.readerMap, key)
		}
		b.readerMu.Unlock()
	}()

	logs, err := reader.Read()
	if err != nil {
		b.logger.Errorf("[BuildEventHandler] error reading build logs for %s: %v", build.Tag(), err)
		// Fail closed: the log stream is where this build's verdict comes
		// from, so no stream means no verdict, and answering SUCCESS would
		// be a guess. In practice this is the cancel window (cancelBuild
		// closed the reader between registering it and reading it) — the
		// caller's SetStatus(ERROR) is then rejected by the entity's CAS
		// because CANCELLED is already decided, so cancel semantics stand.
		return err
	}

	for log := range logs {
		build.AddLog(log)
		if log.Level() == buildloglevel.Error {
			err = errors.New(log.Message())
		}
	}

	return err
}

func (b *BuildEventHandler) cancelBuild(event *event.CancelBuildEvent) {
	build := event.Build()
	// Atomic compare-and-set via the entity (see Build.SetStatus): the old
	// separate CanChangeTo check raced buildImage's own transitions on the
	// same shared entity.
	if err := build.SetStatus(buildstatus.CANCELLED); err != nil {
		return
	}
	if err := b.buildRepo.Commit(build); err != nil {
		return
	}

	key := b.buildKey(build)
	b.readerMu.Lock()
	reader := b.readerMap[key]
	delete(b.readerMap, key)
	b.readerMu.Unlock()

	if reader == nil {
		return
	}
	reader.Close()
}
