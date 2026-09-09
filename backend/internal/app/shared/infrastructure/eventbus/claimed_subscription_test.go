package eventbus_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// claimEvt is the minimal event shape: Topic for the bus, ID for the claim.
type claimEvt struct{ id string }

func (e *claimEvt) Topic() string { return "test:claimed" }
func (e *claimEvt) ID() string    { return e.id }

// fakeEventStore is an in-memory TryClaimEvent/ReleaseEvent with call
// recording — the claim semantics (SetNX/Del) are what the retry protocol is
// built on, so the fake reproduces them exactly.
type fakeEventStore struct {
	mu       sync.Mutex
	claims   map[string]bool
	claimErr error
	relErr   error

	claimCalls   []string
	releaseCalls []string
}

func newFakeEventStore() *fakeEventStore {
	return &fakeEventStore{claims: map[string]bool{}}
}

func (s *fakeEventStore) TryClaimEvent(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claimCalls = append(s.claimCalls, id)
	if s.claimErr != nil {
		return false, s.claimErr
	}
	if s.claims[id] {
		return false, nil
	}
	s.claims[id] = true
	return true, nil
}

func (s *fakeEventStore) ReleaseEvent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.releaseCalls = append(s.releaseCalls, id)
	if s.relErr != nil {
		return s.relErr
	}
	delete(s.claims, id)
	return nil
}

func (s *fakeEventStore) snapshot() (claims, releases []string, held map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	held = map[string]bool{}
	for k, v := range s.claims {
		held[k] = v
	}
	return append([]string{}, s.claimCalls...), append([]string{}, s.releaseCalls...), held
}

// syncOpts makes the whole retry chain run inline on the handler goroutine,
// so the real bus's WaitAsync is a complete barrier over all attempts.
func syncOpts() eventbus.ClaimedOptions[*claimEvt] {
	return eventbus.ClaimedOptions[*claimEvt]{
		Backoff:       func(int) time.Duration { return 0 },
		ScheduleRetry: func(_ time.Duration, f func()) { f() },
	}
}

// subscribe wires a claimed handler over a REAL event bus. failures = how
// many leading calls return an error (-1 = always fail).
func subscribe(t *testing.T, store *fakeEventStore, failures int, opts eventbus.ClaimedOptions[*claimEvt]) (sharedtypes.EventBus, *int) {
	t.Helper()
	bus := eventbus.NewEventBus(sharedlogger.NewMockLogger())
	calls := new(int)
	var mu sync.Mutex
	err := eventbus.SubscribeClaimedAsync(bus, store, sharedlogger.NewMockLogger(), false,
		func(e *claimEvt) error {
			mu.Lock()
			defer mu.Unlock()
			*calls++
			if failures < 0 || *calls <= failures {
				return errors.New("boom")
			}
			return nil
		}, opts)
	require.NoError(t, err)
	return bus, calls
}

func TestClaimed_FirstAttemptSucceeds_ClaimKeptNoRelease(t *testing.T) {
	store := newFakeEventStore()
	bus, calls := subscribe(t, store, 0, syncOpts())

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	claims, releases, held := store.snapshot()
	assert.Equal(t, 1, *calls)
	assert.Equal(t, []string{"e1"}, claims)
	assert.Empty(t, releases, "a successful handler must keep its claim")
	assert.True(t, held["e1"], "the claim stays to dedupe redeliveries")
}

func TestClaimed_DuplicateEventIsDropped(t *testing.T) {
	store := newFakeEventStore()
	bus, calls := subscribe(t, store, 0, syncOpts())

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()
	bus.Publish(&claimEvt{id: "e1"}) // redelivery of the same event
	bus.WaitAsync()

	assert.Equal(t, 1, *calls, "the second delivery must not reach the handler")
}

func TestClaimed_ClaimErrorDropsWithoutRunning(t *testing.T) {
	store := newFakeEventStore()
	store.claimErr = errors.New("redis down")
	bus, calls := subscribe(t, store, 0, syncOpts())

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	_, releases, _ := store.snapshot()
	assert.Equal(t, 0, *calls, "no claim, no execution")
	assert.Empty(t, releases)
}

func TestClaimed_FailureReleasesThenRetrySucceeds(t *testing.T) {
	store := newFakeEventStore()
	bus, calls := subscribe(t, store, 1, syncOpts()) // fail once, then succeed

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	claims, releases, held := store.snapshot()
	assert.Equal(t, 2, *calls, "one failure + one successful retry")
	assert.Equal(t, []string{"e1", "e1"}, claims, "the retry re-claims")
	assert.Equal(t, []string{"e1"}, releases, "only the failed attempt releases")
	assert.True(t, held["e1"], "the successful retry keeps the claim")
}

func TestClaimed_MaxAttemptsExhausted_GivesUpWithClaimReleased(t *testing.T) {
	store := newFakeEventStore()
	opts := syncOpts()
	opts.MaxAttempts = 3
	bus, calls := subscribe(t, store, -1, opts) // always fails

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	claims, releases, held := store.snapshot()
	assert.Equal(t, 3, *calls, "first attempt + two retries")
	assert.Len(t, claims, 3)
	assert.Len(t, releases, 3, "every failure releases, including the last")
	assert.False(t, held["e1"], "after giving up the claim is NOT held — a redelivery may retry")
}

func TestClaimed_ReleaseFailureStopsRetrying(t *testing.T) {
	store := newFakeEventStore()
	store.relErr = errors.New("redis down")
	bus, calls := subscribe(t, store, -1, syncOpts())

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	_, releases, held := store.snapshot()
	assert.Equal(t, 1, *calls, "cannot release ⇒ retrying would hit our own claim; stop")
	assert.Len(t, releases, 1)
	assert.True(t, held["e1"], "the stuck claim stays until TTL")
}

func TestClaimed_CustomKeyUsedForBothClaimAndRelease(t *testing.T) {
	store := newFakeEventStore()
	opts := syncOpts()
	opts.Key = func(e *claimEvt) string { return "prefixed-" + e.ID() }
	opts.MaxAttempts = 2
	bus, calls := subscribe(t, store, -1, opts)

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	claims, releases, _ := store.snapshot()
	assert.Equal(t, 2, *calls)
	for _, c := range claims {
		assert.Equal(t, "prefixed-e1", c)
	}
	for _, r := range releases {
		assert.Equal(t, "prefixed-e1", r)
	}
}

// TestClaimed_DefaultScheduleStaysInsideWaitAsyncBarrier exercises the
// PRODUCTION retry scheduler (ScheduleRetry deliberately nil): the retry must
// complete before WaitAsync returns, because the BDD reset seam uses
// WaitAsync as its "no handler is still running" barrier before wiping state.
// Under the old time.AfterFunc default this test fails — WaitAsync returns
// while the retry is still pending and calls is 1, the ghost-write window.
func TestClaimed_DefaultScheduleStaysInsideWaitAsyncBarrier(t *testing.T) {
	store := newFakeEventStore()
	opts := eventbus.ClaimedOptions[*claimEvt]{
		Backoff: func(int) time.Duration { return time.Millisecond },
		// ScheduleRetry deliberately nil: the production default is under test.
	}
	bus, calls := subscribe(t, store, 1, opts) // fail once, then succeed

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	assert.Equal(t, 2, *calls, "WaitAsync must cover the default-scheduled retry")
}

func TestClaimed_TransactionalSubscriptionSmoke(t *testing.T) {
	store := newFakeEventStore()
	bus := eventbus.NewEventBus(sharedlogger.NewMockLogger())
	calls := 0
	err := eventbus.SubscribeClaimedAsync(bus, store, sharedlogger.NewMockLogger(), true,
		func(e *claimEvt) error { calls++; return nil }, syncOpts())
	require.NoError(t, err)

	bus.Publish(&claimEvt{id: "t1"})
	bus.WaitAsync()
	assert.Equal(t, 1, calls)
}

// ── SubscribeRetriedAsync (no claim) ─────────────────────────────────────────

// subscribeRetried wires a retried (claimless) handler over a real bus.
func subscribeRetried(t *testing.T, failures int, opts eventbus.RetryOptions) (sharedtypes.EventBus, *int) {
	t.Helper()
	bus := eventbus.NewEventBus(sharedlogger.NewMockLogger())
	calls := new(int)
	var mu sync.Mutex
	err := eventbus.SubscribeRetriedAsync(bus, sharedlogger.NewMockLogger(), false,
		func(e *claimEvt) string { return e.ID() },
		func(e *claimEvt) error {
			mu.Lock()
			defer mu.Unlock()
			*calls++
			if failures < 0 || *calls <= failures {
				return errors.New("boom")
			}
			return nil
		}, opts)
	require.NoError(t, err)
	return bus, calls
}

func syncRetryOpts() eventbus.RetryOptions {
	return eventbus.RetryOptions{
		Backoff:       func(int) time.Duration { return 0 },
		ScheduleRetry: func(_ time.Duration, f func()) { f() },
	}
}

func TestRetried_FailureRetriesUntilSuccess(t *testing.T) {
	bus, calls := subscribeRetried(t, 1, syncRetryOpts()) // fail once, then succeed

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	assert.Equal(t, 2, *calls, "one failure + one successful retry")
}

func TestRetried_MaxAttemptsExhausted(t *testing.T) {
	opts := syncRetryOpts()
	opts.MaxAttempts = 3
	bus, calls := subscribeRetried(t, -1, opts) // always fails

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	assert.Equal(t, 3, *calls, "first attempt + two retries, then give up")
}

// TestRetried_EveryDeliveryRuns pins the defining difference from the claimed
// variant: NO dedupe. The same event ID delivered twice runs the handler for
// both deliveries — the rejudge rollup depends on seeing every judgeupdated.
func TestRetried_EveryDeliveryRuns(t *testing.T) {
	bus, calls := subscribeRetried(t, 0, syncRetryOpts()) // never fails

	bus.Publish(&claimEvt{id: "same-id"})
	bus.WaitAsync()
	bus.Publish(&claimEvt{id: "same-id"})
	bus.WaitAsync()

	assert.Equal(t, 2, *calls, "a retried (claimless) subscription must run for every delivery")
}

// TestClaimed_DefaultBackoffIsTwoThenFourSeconds pins the production backoff
// formula ((retry+1) * 2s), which every other test overrides for speed — the
// delays are captured through an injected scheduler that fires immediately,
// so nothing actually sleeps.
func TestClaimed_DefaultBackoffIsTwoThenFourSeconds(t *testing.T) {
	store := newFakeEventStore()
	var delays []time.Duration
	opts := eventbus.ClaimedOptions[*claimEvt]{
		// Backoff deliberately nil: the production default is under test.
		ScheduleRetry: func(d time.Duration, f func()) {
			delays = append(delays, d)
			f()
		},
	}
	bus, calls := subscribe(t, store, -1, opts) // always fails

	bus.Publish(&claimEvt{id: "e1"})
	bus.WaitAsync()

	assert.Equal(t, 3, *calls, "default MaxAttempts")
	assert.Equal(t, []time.Duration{2 * time.Second, 4 * time.Second}, delays,
		"the default backoff must be (retry+1)*2s")
}
