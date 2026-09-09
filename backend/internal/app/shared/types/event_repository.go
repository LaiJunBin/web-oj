package sharedtypes

type EventRepository interface {
	TryClaimEvent(eventID string) (bool, error)
	// ReleaseEvent deletes a claim THIS process took, so the event can be
	// attempted again (the retry path of SubscribeClaimedAsync). Only call
	// it for handlers whose work is idempotent: releasing a claim invites a
	// re-execution.
	ReleaseEvent(eventID string) error
}
