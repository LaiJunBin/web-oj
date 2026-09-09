package query

import (
	"time"

	"judge-backend/internal/app/judge_domain/worker/application/query"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	"judge-backend/internal/app/judge_domain/worker/domain/provider"
)

type JudgeQueueQuery struct {
	stats provider.QueueStats
	// now is injected so the elapsed-wait calculation is testable without
	// sleeping.
	now func() time.Time
}

func NewJudgeQueueQuery(
	stats provider.QueueStats,
) query.JudgeQueueQuery {
	return &JudgeQueueQuery{
		stats: stats,
		now:   time.Now,
	}
}

func (q *JudgeQueueQuery) GetJudgeQueueStatus() (*dto.JudgeQueueStatus, error) {
	snapshot := q.stats.Snapshot()

	depths := make(map[string]int32, len(snapshot.DepthsByJudgeType))
	for judgeType, depth := range snapshot.DepthsByJudgeType {
		depths[judgeType] = int32(depth)
	}

	status := &dto.JudgeQueueStatus{
		QueueDepths:     depths,
		ActiveJudgments: int32(snapshot.ActiveJudgments),
	}

	if snapshot.OldestQueuedAt != nil {
		waited := q.now().Sub(*snapshot.OldestQueuedAt).Seconds()
		// A judgment created a hair in the future (clock adjustment, or a
		// judgment queued in the same instant it was read) would otherwise be
		// reported as a negative wait, which reads as a broken dashboard rather
		// than as the "just arrived" it actually is.
		if waited < 0 {
			waited = 0
		}
		status.OldestQueuedSeconds = &waited
	}

	return status, nil
}
