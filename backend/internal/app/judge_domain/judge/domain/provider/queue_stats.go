package provider

import "time"

// QueueSnapshot is the judge pipeline's load at one instant, in published
// language: judge types are plain strings, not the judge subdomain's judgetype
// values.
type QueueSnapshot struct {
	// DepthsByJudgeType is the number of judgments waiting per judge type.
	// Every judge type the pipeline serves is present, including the ones
	// sitting at zero, so a consumer never has to know the set of types to
	// render a complete picture.
	DepthsByJudgeType map[string]int

	// ActiveJudgments is how many judgments are executing right now.
	ActiveJudgments int

	// OldestQueuedAt is when the longest-waiting judgment across all queues
	// was created, or nil when nothing is queued.
	OldestQueuedAt *time.Time
}

// QueueStats is the judge subdomain's published queue-observability
// capability. The judge infrastructure implements it; downstream contexts
// consume it through their own anti-corruption adapters.
//
// Snapshot is deliberately not atomic across queues: each queue is read under
// its own lock, so a judgment can move between two of them mid-read. Making it
// atomic would mean holding every queue's lock at once — putting a monitoring
// read on the critical path of the pipeline it is monitoring. For a status
// display, a reading that is a few microseconds stale in one field is not worth
// that.
type QueueStats interface {
	Snapshot() QueueSnapshot
}
