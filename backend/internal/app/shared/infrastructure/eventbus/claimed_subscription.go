package eventbus

import (
	"errors"
	"time"

	sharedtypes "judge-backend/internal/app/shared/types"
)

// defaultRetryMaxAttempts counts the FIRST attempt too: 3 means one
// original try plus two retries.
const defaultRetryMaxAttempts = 3

// errAbortRetry signals that an attempt already logged its own terminal
// outcome and must neither be retried nor logged again by the retry core
// (a failed claim, or a release that could not be undone).
var errAbortRetry = errors.New("eventbus: abort retry")

// RetryOptions tunes the bounded in-process retry shared by
// SubscribeRetriedAsync and SubscribeClaimedAsync. The zero value is the
// production configuration; the function fields exist so tests can make the
// retry chain instant.
type RetryOptions struct {
	// MaxAttempts caps how many times the handler runs, first attempt
	// included. <= 0 means the default of 3.
	MaxAttempts int
	// Backoff returns the delay before retry number retry (0-based, i.e.
	// Backoff(0) precedes the second attempt). Nil means (retry+1) * 2s.
	Backoff func(retry int) time.Duration
	// ScheduleRetry runs f after d. Nil means a blocking sleep on the
	// handler's own goroutine, keeping the whole retry chain inside
	// WaitAsync's barrier. Tests inject a zero-delay version so suites do
	// not sleep out real backoff.
	ScheduleRetry func(d time.Duration, f func())
	// LogPrefix labels the retry/give-up log lines; empty means
	// "[RetriedSubscription]".
	LogPrefix string
	// GiveUpNote is appended to the exhausted-attempts log line (the claimed
	// variant uses it to say the claim was released).
	GiveUpNote string
}

// normalized is the single place the retry defaults are defined.
func (o RetryOptions) normalized() RetryOptions {
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = defaultRetryMaxAttempts
	}
	if o.Backoff == nil {
		o.Backoff = func(retry int) time.Duration { return time.Duration(retry+1) * 2 * time.Second }
	}
	if o.ScheduleRetry == nil {
		// Block on the handler's own goroutine instead of time.AfterFunc: the
		// whole retry chain then stays inside the goroutine the bus opened
		// for this event, so WaitAsync (which waits for handler goroutines to
		// return) is a complete barrier over every attempt. time.AfterFunc
		// ran retries OUTSIDE that barrier — the BDD reset seam could declare
		// the bus drained while a retry was still pending, and the retry then
		// wrote into the next scenario's freshly reset database.
		o.ScheduleRetry = func(d time.Duration, f func()) { time.Sleep(d); f() }
	}
	if o.LogPrefix == "" {
		o.LogPrefix = "[RetriedSubscription]"
	}
	return o
}

// SubscribeRetriedAsync subscribes a handler that must see EVERY delivery of
// its event — NO claim, NO dedupe — but whose transient failures retry
// in-process with backoff. The rejudge rollup is the archetype: one judgment
// publishes many judgeupdated events and each carries state, so a claim
// keyed by anything would swallow deliveries it must not; what it needs is
// only that a single Mongo hiccup does not permanently strand the rejudge in
// RUNNING.
//
// The handler must be idempotent per event: a retry re-runs it from the top.
// describe(e) names the event in log lines. Retries are in-process only and
// block on the handler's goroutine (see RetryOptions.ScheduleRetry).
func SubscribeRetriedAsync[E any](
	bus sharedtypes.EventBus,
	logger sharedtypes.Logger,
	transactional bool,
	describe func(e E) string,
	handler func(e E) error,
	opts RetryOptions,
) error {
	o := opts.normalized()

	var attempt func(e E, n int)
	attempt = func(e E, n int) {
		err := handler(e)
		if err == nil || errors.Is(err, errAbortRetry) {
			return
		}
		if n+1 >= o.MaxAttempts {
			logger.Errorf("%s Handler failed for event %s after %d attempts, giving up%s: %v",
				o.LogPrefix, describe(e), o.MaxAttempts, o.GiveUpNote, err)
			return
		}
		delay := o.Backoff(n)
		logger.Warnf("%s Handler failed for event %s (attempt %d/%d), retrying in %s: %v",
			o.LogPrefix, describe(e), n+1, o.MaxAttempts, delay, err)
		o.ScheduleRetry(delay, func() { attempt(e, n+1) })
	}

	// Never unsubscribed, and it must stay that way unless a registry is
	// added: every subscription made through here wraps its handler in THIS
	// one func literal, so all of them share a code pointer — the identity
	// the library would match on — exactly the trap that broke SSE teardown
	// (see Eventbus.Subscribe).
	return bus.SubscribeAsync(func(e E) { attempt(e, 0) }, transactional)
}

// ClaimedOptions tunes one claimed subscription. The zero value is the
// production configuration.
type ClaimedOptions[E any] struct {
	// Key derives the claim key from the event. Nil means the event's own
	// ID() — override it only when the raw ID would collide with another
	// handler's claims for the same event (the contest scoreboard handler
	// prefixes its key for exactly that reason).
	Key func(e E) string
	// MaxAttempts caps how many times the handler runs, first attempt
	// included. <= 0 means the default of 3.
	MaxAttempts int
	// Backoff returns the delay before retry number retry (0-based, i.e.
	// Backoff(0) precedes the second attempt). Nil means (retry+1) * 2s.
	Backoff func(retry int) time.Duration
	// ScheduleRetry runs f after d. Nil means a blocking sleep on the
	// handler's own goroutine, keeping the whole retry chain inside
	// WaitAsync's barrier. Tests inject a zero-delay version so suites do
	// not sleep out real backoff.
	ScheduleRetry func(d time.Duration, f func())
}

// SubscribeClaimedAsync subscribes an IDEMPOTENT handler with the claim /
// release / retry protocol that used to be hand-rolled (claim only) inside
// every handler:
//
//	claim the event  → run the handler
//	handler nil      → done, the claim stays and dedupes redeliveries
//	handler error    → release the claim, wait, try again (bounded)
//
// This closes the silent-drop for idempotent handlers: before, the claim was
// taken BEFORE the side effect and never released, so one transient failure
// (a Mongo hiccup during a cancel, say) silently discarded work the API had
// already confirmed. Now the claim is only kept once the handler succeeded.
//
// The contract, in exchange, is IDEMPOTENCY: a released claim invites
// re-execution, so the handler must tolerate running again (every state
// transition reachable from here goes through a conditional repository
// update, which is what makes a second run harmless). Creation-style
// handlers that mint new IDs per run must NOT use this; they keep their own
// claim-only protocol until their IDs are derivable from the event, which is
// a deferred roadmap item. Handlers that must see EVERY delivery
// (no dedupe at all) use SubscribeRetriedAsync instead.
//
// Retries are in-process only: if the process dies, the pending retry dies
// with it — an accepted boundary. The default retry blocks on the
// handler's own goroutine, so WaitAsync covers the whole chain; the costs are
// that a failing handler's goroutine lives for the backoff total (~6s at the
// defaults), and — for a transactional subscription — that the handler lock
// is acquired on the PUBLISHER's goroutine inside Publish() (the bus unlocks
// its own mutex, takes the handler lock, then spawns the handler goroutine
// which releases it on completion): while a retry chain is sleeping, the next
// publisher of a same-type event BLOCKS inside Publish(), it does not merely
// queue. The BDD reset seam's 10s WaitAsync budget can expire under a
// persistently failing handler — that 504 is deliberate fail-loud, pointing
// at the real failure instead of letting a late retry ghost-write into the
// next scenario.
func SubscribeClaimedAsync[E interface{ ID() string }](
	bus sharedtypes.EventBus,
	repo sharedtypes.EventRepository,
	logger sharedtypes.Logger,
	transactional bool,
	handler func(e E) error,
	opts ClaimedOptions[E],
) error {
	keyFn := opts.Key
	if keyFn == nil {
		keyFn = func(e E) string { return e.ID() }
	}

	// The claim/release protocol wraps the caller's handler; the bounded
	// retry itself is the shared SubscribeRetriedAsync core.
	wrapped := func(e E) error {
		key := keyFn(e)

		claimed, err := repo.TryClaimEvent(key)
		if err != nil {
			// Same posture as every handler had before: a Redis failure
			// drops the event (we cannot even tell whether it is a
			// duplicate).
			logger.Errorf("[ClaimedSubscription] Failed to claim event %s: %v", key, err)
			return errAbortRetry
		}
		if !claimed {
			logger.Debugf("[ClaimedSubscription] Event %s already processed, dropping", key)
			return nil
		}

		if err := handler(e); err == nil {
			return nil // success — the claim stays and dedupes redeliveries
		} else if releaseErr := repo.ReleaseEvent(key); releaseErr != nil {
			// The claim is stuck until its TTL expires and nothing will
			// retry within this process — say so instead of retrying into a
			// claim we know is still held.
			logger.Errorf("[ClaimedSubscription] Handler failed for event %s (%v) and the claim could not be released (%v); giving up in this process, the claim expires by TTL", key, err, releaseErr)
			return errAbortRetry
		} else {
			return err // claim released — retryable
		}
	}

	return SubscribeRetriedAsync(bus, logger, transactional, keyFn, wrapped, RetryOptions{
		MaxAttempts:   opts.MaxAttempts,
		Backoff:       opts.Backoff,
		ScheduleRetry: opts.ScheduleRetry,
		LogPrefix:     "[ClaimedSubscription]",
		GiveUpNote:    " (claim released, a redelivery may retry)",
	})
}
