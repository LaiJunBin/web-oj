package acl

import (
	"errors"
	"testing"
	"time"

	judgeprovider "judge-backend/internal/app/judge_domain/judge/domain/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Hand-written fakes: the contract under test is the TRANSLATION between
// the two providers, so the upstream is faked at the judge contract exactly
// as a downstream consumer sees it.

type fakeJudgeHandle struct {
	id        string
	status    string
	order     []string
	createdAt time.Time
	lastExec  *time.Time
	logs      []judgeprovider.WorkerLog
	started   int
	stopped   int
}

func (f *fakeJudgeHandle) ID() string                       { return f.id }
func (f *fakeJudgeHandle) Status() string                   { return f.status }
func (f *fakeJudgeHandle) JudgeOrder() []string             { return f.order }
func (f *fakeJudgeHandle) CreatedAt() time.Time             { return f.createdAt }
func (f *fakeJudgeHandle) LastExecutionAt() *time.Time      { return f.lastExec }
func (f *fakeJudgeHandle) Logs() []judgeprovider.WorkerLog  { return f.logs }
func (f *fakeJudgeHandle) Start()                           { f.started++ }
func (f *fakeJudgeHandle) Stop()                            { f.stopped++ }

type fakeJudgePool struct {
	handles   map[string]*fakeJudgeHandle
	addedWith [][]string
	addErr    error
	deleted   []string
}

func (f *fakeJudgePool) List() []judgeprovider.WorkerHandle {
	out := make([]judgeprovider.WorkerHandle, 0, len(f.handles))
	for _, h := range f.handles {
		out = append(out, h)
	}
	return out
}

func (f *fakeJudgePool) Get(id string) judgeprovider.WorkerHandle {
	h, ok := f.handles[id]
	if !ok {
		return nil // typed-nil trap: must return untyped nil interface
	}
	return h
}

func (f *fakeJudgePool) Add(judgeOrder []string) error {
	f.addedWith = append(f.addedWith, judgeOrder)
	return f.addErr
}

func (f *fakeJudgePool) Delete(id string) { f.deleted = append(f.deleted, id) }

func TestJudgeWorkerPoolACL_DelegatesAndTranslates(t *testing.T) {
	now := time.Now()
	upstream := &fakeJudgePool{handles: map[string]*fakeJudgeHandle{
		"w1": {
			id: "w1", status: "Idle", order: []string{"JUDGE", "REJUDGE"},
			createdAt: now, lastExec: &now,
			logs: []judgeprovider.WorkerLog{{Text: "hello", Time: now}},
		},
	}}
	pool := NewJudgeWorkerPool(upstream)

	handles := pool.List()
	require.Len(t, handles, 1)
	h := handles[0]
	assert.Equal(t, "w1", h.ID())
	assert.Equal(t, "Idle", h.Status())
	assert.Equal(t, []string{"JUDGE", "REJUDGE"}, h.JudgeOrder())
	assert.Equal(t, now, h.CreatedAt())
	require.NotNil(t, h.LastExecutionAt())

	logs := h.Logs()
	require.Len(t, logs, 1)
	assert.Equal(t, "hello", logs[0].Text)
	assert.Equal(t, now, logs[0].Time)

	h.Start()
	h.Stop()
	assert.Equal(t, 1, upstream.handles["w1"].started)
	assert.Equal(t, 1, upstream.handles["w1"].stopped)

	require.NoError(t, pool.Add([]string{"JUDGE"}))
	assert.Equal(t, [][]string{{"JUDGE"}}, upstream.addedWith)

	pool.Delete("w1")
	assert.Equal(t, []string{"w1"}, upstream.deleted)
}

// TestJudgeWorkerPoolACL_MissingWorkerIsUntypedNil pins the classic
// interface-nil trap: the upstream returning "no such worker" must surface
// as a true nil from the ACL, not a non-nil interface wrapping a nil
// pointer that every caller's `!= nil` check would mistake for a worker.
func TestJudgeWorkerPoolACL_MissingWorkerIsUntypedNil(t *testing.T) {
	pool := NewJudgeWorkerPool(&fakeJudgePool{handles: map[string]*fakeJudgeHandle{}})
	assert.Nil(t, pool.Get("nope"))
}

func TestJudgeWorkerPoolACL_AddErrorPassesThrough(t *testing.T) {
	upstream := &fakeJudgePool{addErr: errors.New("invalid judge order: NOPE")}
	pool := NewJudgeWorkerPool(upstream)
	assert.ErrorContains(t, pool.Add([]string{"NOPE"}), "invalid judge order")
}
