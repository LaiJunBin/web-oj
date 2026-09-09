package queue_test

import (
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	logger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Depth and Len answer different questions, and the admin status endpoint needs
// the one Len does not give: Len counts non-empty per-project sub-queues, so ten
// judgments queued against one project read as 1.
func TestDepthCountsJudgmentsNotProjects(t *testing.T) {
	priorityStrategy := prioritystrategies.NewDefault()
	jq := queue.NewJudgeQueue("JUDGE", logger.NewMockLogger())

	base := time.Now()
	jq.Push(newJudgment("p1-a", 1, base), priorityStrategy)
	jq.Push(newJudgment("p1-b", 1, base.Add(time.Second)), priorityStrategy)
	jq.Push(newJudgment("p1-c", 1, base.Add(2*time.Second)), priorityStrategy)

	assert.Equal(t, 1, jq.Len(), "one project has judgments queued")
	assert.Equal(t, 3, jq.Depth(), "three judgments are waiting")

	jq.Push(newJudgment("p2-a", 2, base.Add(3*time.Second)), priorityStrategy)
	assert.Equal(t, 2, jq.Len())
	assert.Equal(t, 4, jq.Depth())

	require.NotNil(t, jq.Pop())
	assert.Equal(t, 3, jq.Depth(), "popping one judgment lowers the depth by one")
}

func TestDepthIsZeroOnAFreshQueue(t *testing.T) {
	jq := queue.NewJudgeQueue("JUDGE", logger.NewMockLogger())
	assert.Equal(t, 0, jq.Depth())
}

func TestOldestQueuedAtIsAbsentWhenEmpty(t *testing.T) {
	jq := queue.NewJudgeQueue("JUDGE", logger.NewMockLogger())

	_, ok := jq.OldestQueuedAt()
	assert.False(t, ok, "an empty queue has nothing waiting, which is not the same as something that has waited zero seconds")
}

// The whole reason OldestQueuedAt scans instead of peeking. Within one project
// the order is the priority strategy, not chronological, so the oldest judgment
// can sit behind a newer one; reading the sub-queue head would report the newer
// one's age and hide exactly the stall this number exists to expose.
func TestOldestQueuedAtFindsAJudgmentBuriedBehindAHigherPriorityOne(t *testing.T) {
	reverseTime := &reverseTimeStrategy{}
	jq := queue.NewJudgeQueue("JUDGE", logger.NewMockLogger())

	oldest := time.Now().Add(-30 * time.Minute)
	newest := time.Now()

	// reverseTimeStrategy orders newest-first, so the recent judgment becomes
	// the sub-queue head and the half-hour-old one is stuck behind it.
	jq.Push(newJudgment("stalled", 1, oldest), reverseTime)
	jq.Push(newJudgment("recent", 1, newest), reverseTime)

	queuedAt, ok := jq.OldestQueuedAt()
	require.True(t, ok)
	assert.WithinDuration(t, oldest, queuedAt, time.Millisecond,
		"the longest-waiting judgment is the stalled one, not whichever the strategy put on top")
}

func TestOldestQueuedAtSpansProjects(t *testing.T) {
	priorityStrategy := prioritystrategies.NewDefault()
	jq := queue.NewJudgeQueue("JUDGE", logger.NewMockLogger())

	oldest := time.Now().Add(-10 * time.Minute)
	jq.Push(newJudgment("p2-old", 2, oldest), priorityStrategy)
	jq.Push(newJudgment("p1-new", 1, time.Now()), priorityStrategy)

	queuedAt, ok := jq.OldestQueuedAt()
	require.True(t, ok)
	assert.WithinDuration(t, oldest, queuedAt, time.Millisecond)
}

func TestOldestQueuedAtClearsAsTheQueueDrains(t *testing.T) {
	priorityStrategy := prioritystrategies.NewDefault()
	jq := queue.NewJudgeQueue("JUDGE", logger.NewMockLogger())

	jq.Push(newJudgment("only", 1, time.Now()), priorityStrategy)
	_, ok := jq.OldestQueuedAt()
	require.True(t, ok)

	require.NotNil(t, jq.Pop())

	_, ok = jq.OldestQueuedAt()
	assert.False(t, ok, "the queue is empty again")
	assert.Equal(t, 0, jq.Depth())
}
