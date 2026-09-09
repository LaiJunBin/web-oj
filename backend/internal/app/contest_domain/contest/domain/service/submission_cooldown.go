package service

// SubmissionCooldown is the gate that spaces out one contestant's submissions
// to one contest project.
//
// It is a gate rather than a "when did they last submit?" lookup on purpose.
// The submission document is written by an ASYNCHRONOUS handler after the
// request has already been answered (judge_domain subscribes to
// ProjectSubmitted via SubscribeAsync), so anything that decides by reading
// submissions back is racing that write: a burst of clicks all read the same
// stale answer and all get through — precisely the case the cooldown exists
// to stop.
//
// Reserving instead closes the window, because claiming the slot and finding
// out whether it was already claimed are one atomic step.
type SubmissionCooldown interface {
	// Reserve claims this user's cooldown slot for the contest project.
	//
	// Returns ok=true when the caller may proceed. When ok=false the slot is
	// already held and retryAfterSeconds says how much longer for — at least
	// 1, never 0, so a client that trusts it does not retry straight into
	// another rejection.
	//
	// A cooldown of zero or less means the contest runs without one; Reserve
	// admits everybody and records nothing.
	Reserve(userID, contestProjectID int32, cooldownSeconds int32) (ok bool, retryAfterSeconds int32, err error)
}
