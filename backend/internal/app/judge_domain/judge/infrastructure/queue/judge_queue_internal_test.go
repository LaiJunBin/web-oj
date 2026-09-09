package queue

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"
	"time"
)

func newInternalTestJudgment(judgmentID string, projectID int32, createdAt time.Time) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project: project.NewProject(&project.ProjectConfig{
			ID: projectID,
		}),
		CreatedAt: createdAt,
	})
	j.SetID(judgmentID)
	return j
}

func TestJudgeQueuePopSkipsStaleEmptyProjectQueue(t *testing.T) {
	logger := logger.NewMockLogger()
	jq := NewJudgeQueue("JUDGE", logger).(*JudgeQueue)
	strategy := prioritystrategies.NewDefault()
	base := time.Now()

	jq.Push(newInternalTestJudgment("valid", 1, base), strategy)

	staleQueueID := jq.nextID
	jq.queueMap[staleQueueID] = &ProjectQueue{
		projectID: 99,
		nextID:    0,
		itemMap:   map[int32]*judgment.Judgment{},
		heap:      sharedtypes.NewHeap([]int32{}, func(i, j int32) bool { return i < j }),
	}
	jq.projectIdMap[99] = staleQueueID
	jq.heap.Push(staleQueueID)

	popped := jq.Pop()
	if popped == nil {
		t.Fatal("expected valid judgment, got nil")
	}
	if popped.ID() != "valid" {
		t.Fatalf("expected valid judgment to be popped, got %s", popped.ID())
	}

	if popped := jq.Pop(); popped != nil {
		t.Fatalf("expected stale queue cleanup pop to return nil, got %s", popped.ID())
	}
	if _, ok := jq.queueMap[staleQueueID]; ok {
		t.Fatalf("expected stale queue %d to be removed", staleQueueID)
	}
}

func TestJudgeQueuePopSkipsMissingJudgmentEntry(t *testing.T) {
	logger := logger.NewMockLogger()
	jq := NewJudgeQueue("JUDGE", logger).(*JudgeQueue)
	strategy := prioritystrategies.NewDefault()
	base := time.Now()

	jq.Push(newInternalTestJudgment("valid", 1, base), strategy)

	staleQueueID := jq.nextID
	staleProjectQueue := &ProjectQueue{
		projectID: 100,
		nextID:    1,
		itemMap:   map[int32]*judgment.Judgment{},
		heap:      sharedtypes.NewHeap([]int32{}, func(i, j int32) bool { return i < j }),
	}
	staleProjectQueue.heap.Push(0)
	jq.queueMap[staleQueueID] = staleProjectQueue
	jq.projectIdMap[100] = staleQueueID
	jq.heap.Push(staleQueueID)

	popped := jq.Pop()
	if popped == nil {
		t.Fatal("expected valid judgment, got nil")
	}
	if popped.ID() != "valid" {
		t.Fatalf("expected valid judgment to be popped, got %s", popped.ID())
	}

	if popped := jq.Pop(); popped != nil {
		t.Fatalf("expected stale queue cleanup pop to return nil, got %s", popped.ID())
	}
	if _, ok := jq.queueMap[staleQueueID]; ok {
		t.Fatalf("expected stale queue %d to be removed", staleQueueID)
	}
}

func TestJudgeQueuePushSkipsMissingJudgmentDuringInnerHeapReorder(t *testing.T) {
	logger := logger.NewMockLogger()
	jq := NewJudgeQueue("JUDGE", logger).(*JudgeQueue)
	strategy := prioritystrategies.NewDefault()
	base := time.Now()

	jq.Push(newInternalTestJudgment("first", 1, base), strategy)

	projectQueue := jq.queueMap[jq.projectIdMap[1]]
	staleIndex := projectQueue.nextID
	projectQueue.heap.Push(staleIndex)
	projectQueue.nextID++

	jq.Push(newInternalTestJudgment("second", 1, base.Add(time.Second)), strategy)

	if popped := jq.Pop(); popped == nil || popped.ID() != "first" {
		if popped == nil {
			t.Fatal("expected first judgment, got nil")
		}
		t.Fatalf("expected first judgment, got %s", popped.ID())
	}

	if popped := jq.Pop(); popped == nil || popped.ID() != "second" {
		if popped == nil {
			t.Fatal("expected second judgment, got nil")
		}
		t.Fatalf("expected second judgment, got %s", popped.ID())
	}
}
