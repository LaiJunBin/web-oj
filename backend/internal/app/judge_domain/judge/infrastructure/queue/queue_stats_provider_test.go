package queue_test

import (
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	domainqueue "judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	logger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pushTo(t *testing.T, provider domainqueue.JudgeQueueProvider, judgeType, judgmentID string, projectID int32, createdAt time.Time) {
	t.Helper()
	q, err := provider.GetQueue(judgeType)
	require.NoError(t, err)
	q.Push(newJudgment(judgmentID, projectID, createdAt), prioritystrategies.NewDefault())
}

// A status page that only listed the judge types with work in them would make a
// drained queue indistinguishable from one that does not exist.
func TestSnapshotReportsEveryJudgeTypeIncludingEmptyOnes(t *testing.T) {
	log := logger.NewMockLogger()
	provider := queue.NewJudgeQueueMap(log)
	stats := queue.NewQueueStatsProvider(provider, log)

	snapshot := stats.Snapshot()

	assert.Equal(t, map[string]int{"JUDGE": 0, "JUDGE_PARTIAL": 0, "REJUDGE": 0}, snapshot.DepthsByJudgeType)
	assert.Nil(t, snapshot.OldestQueuedAt, "nothing is queued, so nothing has been waiting")
}

func TestSnapshotCountsEachJudgeTypeSeparately(t *testing.T) {
	log := logger.NewMockLogger()
	provider := queue.NewJudgeQueueMap(log)
	stats := queue.NewQueueStatsProvider(provider, log)

	base := time.Now()
	pushTo(t, provider, "JUDGE", "j-1", 1, base)
	pushTo(t, provider, "JUDGE", "j-2", 1, base)
	pushTo(t, provider, "REJUDGE", "r-1", 2, base)

	snapshot := stats.Snapshot()

	assert.Equal(t, map[string]int{"JUDGE": 2, "JUDGE_PARTIAL": 0, "REJUDGE": 1}, snapshot.DepthsByJudgeType)
}

// The oldest wait is a property of the pipeline, not of one queue: a judgment
// stuck in REJUDGE while JUDGE keeps turning over is exactly the stall this
// number is for.
func TestSnapshotTakesTheOldestWaitAcrossAllQueues(t *testing.T) {
	log := logger.NewMockLogger()
	provider := queue.NewJudgeQueueMap(log)
	stats := queue.NewQueueStatsProvider(provider, log)

	oldest := time.Now().Add(-45 * time.Minute)
	pushTo(t, provider, "JUDGE", "recent", 1, time.Now())
	pushTo(t, provider, "REJUDGE", "stalled", 2, oldest)

	snapshot := stats.Snapshot()

	require.NotNil(t, snapshot.OldestQueuedAt)
	assert.WithinDuration(t, oldest, *snapshot.OldestQueuedAt, time.Millisecond)
}
