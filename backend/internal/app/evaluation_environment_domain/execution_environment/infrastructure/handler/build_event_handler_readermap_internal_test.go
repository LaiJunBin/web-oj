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
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	logreader "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/log_reader"
	dockermocks "judge-backend/internal/mocks/github.com/docker/docker/client"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// This file is deliberately `package handler` (white-box): the invariant
// under test is about WHICH reader the readerMap holds at each step, and the
// interleaving cannot be sequenced deterministically from outside the
// package.

func waitForReader(t *testing.T, b *BuildEventHandler, key string, cond func(*logreader.LogReader) bool) *logreader.LogReader {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		b.readerMu.Lock()
		r := b.readerMap[key]
		b.readerMu.Unlock()
		if cond(r) {
			return r
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for readerMap[%s] to reach the expected state", key)
	return nil
}

// TestExecute_ConcurrentSameTagBuild_FirstFinishDoesNotOrphanSecond pins the
// guarded delete in execute's defer: two concurrent builds of the same tag
// share one readerMap key, so the second build overwrites the first's
// reader. When the FIRST build then finishes, an unconditional delete would
// remove the SECOND build's reader from the map — leaving that still-running
// build impossible to cancel (cancelBuild would find nothing to close). The
// defer must only delete the entry while it still points at its own reader.
func TestExecute_ConcurrentSameTagBuild_FirstFinishDoesNotOrphanSecond(t *testing.T) {
	fss := sharedmocks.NewFileSystemService(t)
	dockerClient := dockermocks.NewMockAPIClient(t)

	b := &BuildEventHandler{
		client:    dockerClient,
		fss:       fss,
		logger:    sharedlogger.NewMockLogger(),
		readerMap: map[string]*logreader.LogReader{},
	}

	projectPath := "/tmp/orphan-race-project"
	buildA := build.New(&build.BuildConfig{Tag: imagetag.MustNew("orphan-race:latest"), Status: buildstatus.PENDING, Batch: batch.MustNew(1)})
	buildB := build.New(&build.BuildConfig{Tag: imagetag.MustNew("orphan-race:latest"), Status: buildstatus.PENDING, Batch: batch.MustNew(1)})
	key := b.buildKey(buildA)
	require.Equal(t, key, b.buildKey(buildB), "both builds must map to the same readerMap key for this test to mean anything")

	pipeAReader, pipeAWriter := io.Pipe()
	pipeBReader, pipeBWriter := io.Pipe()
	defer pipeBWriter.Close()
	defer pipeAWriter.Close()

	fss.On("Exists", projectPath).Return(true)
	fss.On("TarPath", projectPath).Return(&bytes.Buffer{}, nil)
	dockerClient.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: pipeAReader}, nil).Once()
	dockerClient.On("ImageBuild", mock.Anything, mock.Anything, mock.Anything).
		Return(types.ImageBuildResponse{Body: pipeBReader}, nil).Once()

	options := &buildservice.BuildOptions{}

	aDone := make(chan struct{})
	go func() {
		defer close(aDone)
		_ = b.execute(buildA, projectPath, options)
	}()
	readerA := waitForReader(t, b, key, func(r *logreader.LogReader) bool { return r != nil })

	bDone := make(chan struct{})
	go func() {
		defer close(bDone)
		_ = b.execute(buildB, projectPath, options)
	}()
	readerB := waitForReader(t, b, key, func(r *logreader.LogReader) bool { return r != nil && r != readerA })

	// Build A finishes while B is still running.
	pipeAWriter.Close()
	select {
	case <-aDone:
	case <-time.After(5 * time.Second):
		t.Fatal("build A did not finish after its log stream closed")
	}

	// The key must still hold B's reader — A's deferred cleanup must not
	// have deleted an entry that no longer belongs to it.
	b.readerMu.Lock()
	current := b.readerMap[key]
	b.readerMu.Unlock()
	assert.Equal(t, readerB, current, "build A's cleanup removed build B's reader — B can no longer be cancelled")

	// And closing B's reader (what cancelBuild does) must still unblock B.
	readerB.Close()
	select {
	case <-bDone:
	case <-time.After(5 * time.Second):
		t.Fatal("build B did not finish after its reader was closed")
	}
}
