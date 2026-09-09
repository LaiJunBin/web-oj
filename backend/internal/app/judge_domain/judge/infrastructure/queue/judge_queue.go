package queue

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	sharedtypes "judge-backend/internal/app/shared/types"
	"sync"
	"time"
)

type ProjectQueue struct {
	heap      *sharedtypes.Heap[int32]
	itemMap   map[int32]*judgment.Judgment
	nextID    int32
	projectID int32
}

type JudgeQueue struct {
	// name labels this queue's metrics (the judge type it serves).
	name                       string
	contestPriorityStrategyMap map[int32]prioritystrategies.Strategy
	heap                       *sharedtypes.Heap[int32]
	queueMap                   map[int32]*ProjectQueue
	nextID                     int32
	projectIdMap               map[int32]int32
	logger                     sharedtypes.Logger
	mu                         sync.Mutex
}

func NewJudgeQueue(name string, logger sharedtypes.Logger) queue.JudgeQueue {
	jq := &JudgeQueue{
		name:                       name,
		contestPriorityStrategyMap: make(map[int32]prioritystrategies.Strategy),
		queueMap:                   make(map[int32]*ProjectQueue),
		nextID:                     0,
		projectIdMap:               make(map[int32]int32),
		logger:                     logger,
	}
	// A fresh queue is empty — this also zeroes the gauge when Reset swaps
	// in new queues.
	metrics.SetJudgeQueueDepth(name, 0)
	jq.heap = sharedtypes.NewHeap([]int32{}, func(i, j int32) bool {
		left := jq.peekJudgment(i)
		right := jq.peekJudgment(j)
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}

		return left.CreatedAt().Before(right.CreatedAt())
	})

	jq.logger.Infof("[JudgeQueue] Created new judge queue")
	return jq
}

func (jq *JudgeQueue) Push(j *judgment.Judgment, priorityStrategy prioritystrategies.Strategy) {
	// Resolve sorting fields BEFORE taking the queue lock: some strategies
	// (e.g. TimeSinceLast) issue DB queries here, and holding jq.mu during
	// that I/O would serialize every Push/Pop/Len call on this queue behind
	// DB latency.
	j.SetSortingFields(priorityStrategy.GetExtraFields(j))

	jq.mu.Lock()
	defer jq.mu.Unlock()
	defer jq.updateDepthMetricLocked()

	projectID := j.Project().ID()
	qid, ok := jq.projectIdMap[projectID]
	if !ok {
		jq.projectIdMap[projectID] = jq.nextID
		qid = jq.projectIdMap[projectID]
		jq.contestPriorityStrategyMap[qid] = priorityStrategy
		jq.nextID++
		jq.logger.Infof("[JudgeQueue] Created new queue for project %d with ID %d", projectID, qid)
	}

	if _, ok := jq.queueMap[qid]; !ok {
		jq.queueMap[qid] = &ProjectQueue{
			projectID: projectID,
			nextID:    0,
			itemMap:   make(map[int32]*judgment.Judgment),
			heap: sharedtypes.NewHeap([]int32{}, func(i, j int32) bool {
				left := jq.queueMap[qid].itemMap[i]
				right := jq.queueMap[qid].itemMap[j]
				if left == nil {
					return false
				}
				if right == nil {
					return true
				}

				return jq.contestPriorityStrategyMap[qid].Priority(left, right)
			}),
		}

		jq.heap.Push(qid)
		jq.logger.Debugf("[JudgeQueue] Initialized project queue %d", qid)
	}

	projectQueue := jq.queueMap[qid]
	projectQueue.itemMap[projectQueue.nextID] = j
	projectQueue.heap.Push(projectQueue.nextID)
	projectQueue.nextID++

	jq.logger.Infof("[JudgeQueue] Pushed judgment %s to queue %d (project %d)", j.ID(), qid, projectID)

	if jq.contestPriorityStrategyMap[qid] != priorityStrategy {
		jq.logger.Infof("[JudgeQueue] Updating priority strategy for queue %d", qid)
		jq.contestPriorityStrategyMap[qid] = priorityStrategy
		items := make([]int32, 0)
		for projectQueue.heap.Len() > 0 {
			items = append(items, projectQueue.heap.Pop())
		}
		for _, item := range items {
			projectQueue.heap.Push(item)
		}

		qids := make([]int32, 0)
		for jq.heap.Len() > 0 {
			qids = append(qids, jq.heap.Pop())
		}
		for _, qid := range qids {
			jq.heap.Push(qid)
		}
		jq.logger.Debugf("[JudgeQueue] Reordered queue %d with new priority strategy", qid)
	}
}

func (jq *JudgeQueue) Pop() *judgment.Judgment {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	defer jq.updateDepthMetricLocked()

	for jq.heap.Len() > 0 {
		qid := jq.heap.Pop()
		projectQueue, ok := jq.queueMap[qid]
		if !ok || projectQueue == nil {
			jq.logger.Warnf("[JudgeQueue] Queue %d missing from queueMap, skipping stale entry", qid)
			continue
		}

		if projectQueue.heap.Len() == 0 {
			jq.removeProjectQueue(qid, projectQueue)
			jq.logger.Warnf("[JudgeQueue] Queue %d has no pending judgments, skipping stale entry", qid)
			continue
		}

		judgmentIndex := projectQueue.heap.Pop()
		judgment, ok := projectQueue.itemMap[judgmentIndex]
		if !ok || judgment == nil {
			delete(projectQueue.itemMap, judgmentIndex)
			if projectQueue.heap.Len() > 0 {
				jq.heap.Push(qid)
			} else {
				jq.removeProjectQueue(qid, projectQueue)
			}
			jq.logger.Warnf("[JudgeQueue] Queue %d missing judgment at index %d, skipping stale entry", qid, judgmentIndex)
			continue
		}

		delete(projectQueue.itemMap, judgmentIndex)
		jq.logger.Infof("[JudgeQueue] Popped judgment %s from queue %d", judgment.ID(), qid)

		if projectQueue.heap.Len() > 0 {
			jq.heap.Push(qid)
			jq.logger.Debugf("[JudgeQueue] Queue %d still has %d items", qid, projectQueue.heap.Len())
		} else {
			jq.removeProjectQueue(qid, projectQueue)
		}

		return judgment
	}

	jq.logger.Debugf("[JudgeQueue] Queue is empty, nothing to pop")
	return nil
}

func (jq *JudgeQueue) Len() int {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	return jq.heap.Len()
}

func (jq *JudgeQueue) Depth() int {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	return jq.depthLocked()
}

// OldestQueuedAt scans every sub-queue rather than only their heads — see the
// interface doc comment for why the head of a sub-queue is not necessarily its
// oldest judgment.
func (jq *JudgeQueue) OldestQueuedAt() (time.Time, bool) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	var oldest time.Time
	found := false
	for _, pq := range jq.queueMap {
		if pq == nil {
			continue
		}
		for _, j := range pq.itemMap {
			if j == nil {
				continue
			}
			if !found || j.CreatedAt().Before(oldest) {
				oldest = j.CreatedAt()
				found = true
			}
		}
	}
	return oldest, found
}

// updateDepthMetricLocked publishes the total number of queued judgments.
// Must be called with jq.mu held; Push/Pop defer it right after the unlock
// defer so it runs while the lock is still held.
func (jq *JudgeQueue) updateDepthMetricLocked() {
	metrics.SetJudgeQueueDepth(jq.name, jq.depthLocked())
}

// depthLocked sums the per-project sub-queues (jq.heap.Len() only counts
// projects). Shared by the metric and by Depth so the number the admin API
// reports and the number Grafana plots can never drift apart.
func (jq *JudgeQueue) depthLocked() int {
	depth := 0
	for _, pq := range jq.queueMap {
		if pq != nil && pq.heap != nil {
			depth += pq.heap.Len()
		}
	}
	return depth
}

func (jq *JudgeQueue) peekJudgment(qid int32) *judgment.Judgment {
	projectQueue, ok := jq.queueMap[qid]
	if !ok || projectQueue == nil || projectQueue.heap.Len() == 0 {
		return nil
	}

	judgmentIndex := projectQueue.heap.Peek()
	judgment, ok := projectQueue.itemMap[judgmentIndex]
	if !ok || judgment == nil {
		return nil
	}

	return judgment
}

func (jq *JudgeQueue) removeProjectQueue(qid int32, projectQueue *ProjectQueue) {
	if projectQueue != nil && projectQueue.heap != nil {
		projectQueue.heap.Close()
	}

	projectID := int32(0)
	if projectQueue != nil {
		projectID = projectQueue.projectID
	}

	delete(jq.queueMap, qid)
	delete(jq.projectIdMap, projectID)
	jq.logger.Infof("[JudgeQueue] Removed empty queue %d and closed its heap", qid)
}
