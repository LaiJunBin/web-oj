package dto

// JudgeQueueStatus is the judge pipeline's load as the admin status page reads
// it.
type JudgeQueueStatus struct {
	// QueueDepths is keyed by judge type, and carries an entry for every type
	// the pipeline serves — including those at zero.
	QueueDepths map[string]int32

	ActiveJudgments int32

	// OldestQueuedSeconds is how long the longest-waiting judgment has been
	// waiting, and nil when nothing is queued. Seconds are resolved here rather
	// than by the caller so every reader measures the wait against the same
	// instant.
	OldestQueuedSeconds *float64
}
