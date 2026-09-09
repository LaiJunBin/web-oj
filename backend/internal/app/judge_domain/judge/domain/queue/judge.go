package queue

import (
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
)

type JudgeQueue interface {
	Push(value *judgment.Judgment, priorityStrategy prioritystrategies.Strategy)
	Pop() *judgment.Judgment

	// Len is the number of non-empty per-project sub-queues, NOT the number of
	// queued judgments — a project with ten judgments waiting counts once. It
	// is still a correct emptiness test (zero exactly when nothing is queued),
	// which is all its callers use it for. For "how many judgments are
	// waiting", use Depth.
	Len() int

	// Depth is the number of judgments waiting across every per-project
	// sub-queue. This is the number published as the judge_queue_depth metric.
	Depth() int

	// OldestQueuedAt is the creation time of the longest-waiting judgment in
	// this queue, and false when the queue is empty.
	//
	// Every judgment is examined rather than just the sub-queue heads: the
	// outer heap orders sub-queues by their head's creation time, but WITHIN a
	// sub-queue the order is the project's priority strategy, which is not
	// chronological. The oldest judgment overall can therefore sit anywhere.
	OldestQueuedAt() (time.Time, bool)
}

// JudgeQueueProvider hands out the queue for a judge type. Consumers depend
// on this interface, not on the in-memory map implementation, so the queue
// backend can be swapped (e.g. for a durable one).
//
// Note on durability: the current in-memory implementation loses queued
// judgments on restart; the orphan repair services compensate — at startup
// and then on a periodic sweep.
type JudgeQueueProvider interface {
	GetQueue(judgeType string) (JudgeQueue, error)

	// TryPop atomically pops the next judgment for judgeType, or returns nil
	// if there is none. Workers MUST use this instead of GetQueue+Pop: those
	// are two separate lock acquisitions, and Reset() swapping the queue map
	// in between them can't stop a caller that already holds the pre-reset
	// queue reference from popping a stale judgment anyway. TryPop holds the
	// provider's lock across the whole lookup-and-pop, so it can never
	// observe a queue Reset() is concurrently replacing.
	TryPop(judgeType string) (*judgment.Judgment, error)

	// Reset drops every queued judgment from all queues. Used as a test seam
	// between BDD scenarios. Safe to call concurrently with TryPop (that's
	// the whole point of TryPop over GetQueue+Pop) — but callers still need
	// to separately wait for judgments already popped before Reset to finish
	// executing (see the worker pool's WaitForIdle).
	Reset()
}
