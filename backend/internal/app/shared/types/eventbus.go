package sharedtypes

// Subscription is the handle a synchronous Subscribe returns, and the ONLY
// way to end that subscription. Unsubscribe removes exactly this
// subscription — never a neighbour that happens to look alike — is
// idempotent, and may be called from inside the subscriber's own callback.
// After it returns, at most one delivery that was already in flight
// (snapshotted before the call) can still reach the handler.
type Subscription interface {
	Unsubscribe()
}

type EventBus interface {
	Publish(event Event)

	// Subscribe registers a synchronous handler — a func(E) whose single
	// parameter type E has a Topic() method — that runs inline on the
	// publisher's goroutine, in subscription order. A synchronous handler
	// must not call back into the bus: Publish, Subscribe, SubscribeAsync and
	// WaitAsync all deadlock on the underlying library's lock, which the
	// publisher still holds while the handler runs. Subscription.Unsubscribe
	// is the one bus operation that is safe from inside a handler.
	Subscribe(handler interface{}) (Subscription, error)
	SubscribeAsync(handler interface{}, transactional bool) error

	// WaitAsync blocks until every async handler currently in flight (across
	// all topics, including any further events those handlers themselves
	// publish) has returned. Publish only enqueues an async handler onto its
	// own goroutine and returns immediately, so a caller that needs to know
	// "everything triggered so far has actually finished" — not just "was
	// submitted" — must wait on this explicitly.
	WaitAsync()
}
