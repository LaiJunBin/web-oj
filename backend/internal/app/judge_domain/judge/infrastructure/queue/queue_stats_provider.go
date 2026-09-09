package queue

import (
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	judgeprovider "judge-backend/internal/app/judge_domain/judge/domain/provider"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	sharedtypes "judge-backend/internal/app/shared/types"
)

// QueueStatsProvider implements the judge subdomain's published
// queue-observability contract (domain/provider.QueueStats) over the live
// in-memory queues, the same way WorkerPoolProvider does for the worker pool.
// The translation out of judge-internal vocabulary (judgetype) into published
// language happens here.
type QueueStatsProvider struct {
	queueProvider queue.JudgeQueueProvider
	logger        sharedtypes.Logger
}

func NewQueueStatsProvider(
	queueProvider queue.JudgeQueueProvider,
	logger sharedtypes.Logger,
) judgeprovider.QueueStats {
	return &QueueStatsProvider{
		queueProvider: queueProvider,
		logger:        logger,
	}
}

func (p *QueueStatsProvider) Snapshot() judgeprovider.QueueSnapshot {
	depths := make(map[string]int, len(judgetype.JudgeTypeMap))

	var oldest time.Time
	found := false
	for name := range judgetype.JudgeTypeMap {
		// Every judge type is reported, including empty ones — see the
		// contract's DepthsByJudgeType note.
		depths[name] = 0

		q, err := p.queueProvider.GetQueue(name)
		if err != nil {
			// A judge type with no queue is a wiring bug, not a runtime
			// condition. Reporting it as zero (and logging) keeps the status
			// endpoint answering rather than failing the whole snapshot.
			p.logger.Errorf("[QueueStatsProvider] No queue for judge type %s: %v", name, err)
			continue
		}

		depths[name] = q.Depth()
		if queuedAt, ok := q.OldestQueuedAt(); ok && (!found || queuedAt.Before(oldest)) {
			oldest = queuedAt
			found = true
		}
	}

	snapshot := judgeprovider.QueueSnapshot{
		DepthsByJudgeType: depths,
		ActiveJudgments:   metrics.ActiveJudgmentCount(),
	}
	if found {
		snapshot.OldestQueuedAt = &oldest
	}
	return snapshot
}
