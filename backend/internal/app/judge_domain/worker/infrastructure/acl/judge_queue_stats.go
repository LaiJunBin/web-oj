package acl

import (
	judgeprovider "judge-backend/internal/app/judge_domain/judge/domain/provider"
	workerprovider "judge-backend/internal/app/judge_domain/worker/domain/provider"
)

type judgeQueueStats struct {
	stats judgeprovider.QueueStats
}

// NewJudgeQueueStats adapts the judge subdomain's published queue-stats
// contract into this subdomain's QueueStats port. Like the worker-pool adapter
// alongside it, the two contracts are near-identical today and this is a thin
// translation; it exists so nothing inside this subdomain imports judge.
func NewJudgeQueueStats(stats judgeprovider.QueueStats) workerprovider.QueueStats {
	return &judgeQueueStats{stats: stats}
}

func (a *judgeQueueStats) Snapshot() workerprovider.QueueSnapshot {
	upstream := a.stats.Snapshot()

	// Copied rather than aliased: the map crosses a context boundary here, and
	// a downstream caller mutating it must not be able to reach back into
	// whatever the judge side handed over.
	depths := make(map[string]int, len(upstream.DepthsByJudgeType))
	for judgeType, depth := range upstream.DepthsByJudgeType {
		depths[judgeType] = depth
	}

	return workerprovider.QueueSnapshot{
		DepthsByJudgeType: depths,
		ActiveJudgments:   upstream.ActiveJudgments,
		OldestQueuedAt:    upstream.OldestQueuedAt,
	}
}
