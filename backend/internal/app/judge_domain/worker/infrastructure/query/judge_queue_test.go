package query

import (
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/worker/domain/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A hand-written stub rather than a generated mock: the port has one method and
// every case here is about what the query does with its return value, which
// reads better as a literal than as a chain of expectations.
type stubQueueStats struct {
	snapshot provider.QueueSnapshot
}

func (s *stubQueueStats) Snapshot() provider.QueueSnapshot { return s.snapshot }

func newQueryAt(now time.Time, snapshot provider.QueueSnapshot) *JudgeQueueQuery {
	return &JudgeQueueQuery{
		stats: &stubQueueStats{snapshot: snapshot},
		now:   func() time.Time { return now },
	}
}

func TestGetJudgeQueueStatus_EmptyQueueHasNoOldestWait(t *testing.T) {
	q := newQueryAt(time.Now(), provider.QueueSnapshot{
		DepthsByJudgeType: map[string]int{"JUDGE": 0, "JUDGE_PARTIAL": 0, "REJUDGE": 0},
		ActiveJudgments:   0,
	})

	status, err := q.GetJudgeQueueStatus()

	require.NoError(t, err)
	assert.Nil(t, status.OldestQueuedSeconds,
		"nothing is queued, which must not be reported as something that has waited zero seconds")
	assert.Equal(t, map[string]int32{"JUDGE": 0, "JUDGE_PARTIAL": 0, "REJUDGE": 0}, status.QueueDepths)
	assert.EqualValues(t, 0, status.ActiveJudgments)
}

func TestGetJudgeQueueStatus_ResolvesTheWaitAgainstOneInstant(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 30, 0, time.UTC)
	queuedAt := now.Add(-90 * time.Second)

	q := newQueryAt(now, provider.QueueSnapshot{
		DepthsByJudgeType: map[string]int{"JUDGE": 3, "JUDGE_PARTIAL": 0, "REJUDGE": 1},
		ActiveJudgments:   2,
		OldestQueuedAt:    &queuedAt,
	})

	status, err := q.GetJudgeQueueStatus()

	require.NoError(t, err)
	require.NotNil(t, status.OldestQueuedSeconds)
	assert.InDelta(t, 90, *status.OldestQueuedSeconds, 0.001)
	assert.Equal(t, map[string]int32{"JUDGE": 3, "JUDGE_PARTIAL": 0, "REJUDGE": 1}, status.QueueDepths)
	assert.EqualValues(t, 2, status.ActiveJudgments)
}

// A judgment queued in the same instant it is read, or a clock stepping
// backwards, must not surface as a negative wait — that reads as a broken
// dashboard rather than as "just arrived".
func TestGetJudgeQueueStatus_ClampsAWaitFromTheFuture(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	queuedAt := now.Add(2 * time.Second)

	q := newQueryAt(now, provider.QueueSnapshot{
		DepthsByJudgeType: map[string]int{"JUDGE": 1},
		OldestQueuedAt:    &queuedAt,
	})

	status, err := q.GetJudgeQueueStatus()

	require.NoError(t, err)
	require.NotNil(t, status.OldestQueuedSeconds)
	assert.Equal(t, float64(0), *status.OldestQueuedSeconds)
}
