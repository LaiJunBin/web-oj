package provider

import "time"

// QueueSnapshot is this subdomain's own vocabulary for the judge pipeline's
// load. It mirrors the judge subdomain's published shape today; the ACL absorbs
// any future divergence (see infrastructure/acl).
type QueueSnapshot struct {
	DepthsByJudgeType map[string]int
	ActiveJudgments   int

	// OldestQueuedAt is nil when nothing is queued.
	OldestQueuedAt *time.Time
}

// QueueStats is this subdomain's port onto judge-pipeline queue observability.
type QueueStats interface {
	Snapshot() QueueSnapshot
}
