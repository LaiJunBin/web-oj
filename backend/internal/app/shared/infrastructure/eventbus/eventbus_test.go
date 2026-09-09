package eventbus_test

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
)

// These tests cover the synchronous Subscribe path — the wrapper-side
// subscriber registry that replaced identity-based unsubscription. The async
// path is covered by claimed_subscription_test.go.

type evtA struct{ N int }

func (*evtA) Topic() string { return "test:sync:a" }

type evtB struct{}

func (*evtB) Topic() string { return "test:sync:b" }

func newBus() sharedtypes.EventBus {
	return sharedeventbus.NewEventBus(sharedlogger.NewMockLogger())
}

// recorder stamps every subscriber out of ONE func literal — the shape whose
// shared code pointer defeated the library's identity matching — so these
// tests double as a regression for the original premise.
func recorder(hits *[]string, id string) func(*evtA) {
	return func(*evtA) { *hits = append(*hits, id) }
}

func mustSubscribe(t *testing.T, bus sharedtypes.EventBus, handler interface{}) sharedtypes.Subscription {
	t.Helper()
	sub, err := bus.Subscribe(handler)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if sub == nil {
		t.Fatal("Subscribe returned a nil handle without an error")
	}
	return sub
}

func TestSubscribe_UnsubscribingTheSecondLeavesTheFirstLive(t *testing.T) {
	bus := newBus()
	var hits []string
	mustSubscribe(t, bus, recorder(&hits, "first"))
	second := mustSubscribe(t, bus, recorder(&hits, "second"))

	second.Unsubscribe()
	bus.Publish(&evtA{})

	if !reflect.DeepEqual(hits, []string{"first"}) {
		t.Fatalf("after unsubscribing the second, got %v, want [first]", hits)
	}
}

func TestSubscribe_UnsubscribingTheFirstLeavesTheSecondLive(t *testing.T) {
	bus := newBus()
	var hits []string
	first := mustSubscribe(t, bus, recorder(&hits, "first"))
	mustSubscribe(t, bus, recorder(&hits, "second"))

	first.Unsubscribe()
	bus.Publish(&evtA{})

	if !reflect.DeepEqual(hits, []string{"second"}) {
		t.Fatalf("after unsubscribing the first, got %v, want [second]", hits)
	}
}

// A topic's fan-out is registered with the library once and kept; a second
// registration would deliver every event twice. Draining the topic and
// subscribing again is the path that would expose it.
func TestSubscribe_ResubscribeAfterDrainDeliversExactlyOnce(t *testing.T) {
	bus := newBus()
	var hits []string
	first := mustSubscribe(t, bus, recorder(&hits, "first"))
	first.Unsubscribe()
	mustSubscribe(t, bus, recorder(&hits, "second"))

	bus.Publish(&evtA{})

	if !reflect.DeepEqual(hits, []string{"second"}) {
		t.Fatalf("got %v, want exactly one delivery to the new subscriber", hits)
	}
}

func TestSubscription_UnsubscribeIsIdempotent(t *testing.T) {
	bus := newBus()
	var hits []string
	a := mustSubscribe(t, bus, recorder(&hits, "a"))
	b := mustSubscribe(t, bus, recorder(&hits, "b"))

	a.Unsubscribe()
	a.Unsubscribe()
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.Unsubscribe()
		}()
	}
	wg.Wait()

	bus.Publish(&evtA{})
	if !reflect.DeepEqual(hits, []string{"b"}) {
		t.Fatalf("repeated unsubscribes of a must not touch b: got %v", hits)
	}

	// Draining the list entirely, then unsubscribing again, must be a no-op.
	b.Unsubscribe()
	b.Unsubscribe()
	hits = nil
	bus.Publish(&evtA{})
	if len(hits) != 0 {
		t.Fatalf("publish on a drained topic delivered %v", hits)
	}
}

func TestSubscribe_TopicsAreIsolated(t *testing.T) {
	bus := newBus()
	var hits []string
	mustSubscribe(t, bus, recorder(&hits, "a"))
	mustSubscribe(t, bus, func(*evtB) { hits = append(hits, "b") })

	bus.Publish(&evtA{})

	if !reflect.DeepEqual(hits, []string{"a"}) {
		t.Fatalf("publishing A reached %v, want [a]", hits)
	}
}

// A synchronous handler has already run when Publish returns — no WaitAsync.
// This is what the SSE callback and the test observers in other packages
// rely on, and what distinguishes Subscribe from SubscribeAsync.
func TestSubscribe_HandlerRunsInlineOnThePublisherGoroutine(t *testing.T) {
	bus := newBus()
	got := 0
	mustSubscribe(t, bus, func(e *evtA) { got = e.N })

	bus.Publish(&evtA{N: 7})

	if got != 7 {
		t.Fatalf("handler had not run when Publish returned (got %d)", got)
	}
}

func TestSubscribe_DeliversInSubscriptionOrder(t *testing.T) {
	bus := newBus()
	var hits []string
	mustSubscribe(t, bus, recorder(&hits, "1"))
	mustSubscribe(t, bus, recorder(&hits, "2"))
	mustSubscribe(t, bus, recorder(&hits, "3"))

	bus.Publish(&evtA{})

	if !reflect.DeepEqual(hits, []string{"1", "2", "3"}) {
		t.Fatalf("got %v, want subscription order", hits)
	}
}

// The fan-out snapshots the list and releases the wrapper's lock before
// calling anyone, so a handler may end its own subscription from inside its
// callback — and must not be called again afterwards.
func TestSubscription_SelfUnsubscribeInsideCallbackDoesNotDeadlock(t *testing.T) {
	bus := newBus()
	calls := 0
	var self sharedtypes.Subscription
	self = mustSubscribe(t, bus, func(*evtA) {
		calls++
		self.Unsubscribe()
	})

	done := make(chan struct{})
	go func() {
		bus.Publish(&evtA{})
		bus.Publish(&evtA{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish deadlocked on a handler that unsubscribes itself")
	}
	if calls != 1 {
		t.Fatalf("self-unsubscribed handler ran %d times, want 1", calls)
	}
}

func TestSubscribe_RejectsInvalidHandlers(t *testing.T) {
	bus := newBus()
	for name, handler := range map[string]interface{}{
		"not a func":              "nope",
		"zero-argument func":      func() {},
		"parameter without Topic": func(int) {},
		"nil":                     nil,
	} {
		sub, err := bus.Subscribe(handler)
		if err == nil {
			t.Errorf("%s: expected an error", name)
		}
		if sub != nil {
			t.Errorf("%s: expected a nil handle alongside the error, got %T", name, sub)
		}
	}
}

// Sync subscribers live in the wrapper's registry; async handlers still
// register with the library directly. Both must keep receiving on a shared
// topic.
func TestSubscribe_CoexistsWithAsyncHandlersOnTheSameTopic(t *testing.T) {
	bus := newBus()
	var syncHits, asyncHits atomic.Int32
	if err := bus.SubscribeAsync(func(*evtA) { asyncHits.Add(1) }, false); err != nil {
		t.Fatalf("SubscribeAsync: %v", err)
	}
	mustSubscribe(t, bus, func(*evtA) { syncHits.Add(1) })

	bus.Publish(&evtA{})
	bus.WaitAsync()

	if syncHits.Load() != 1 || asyncHits.Load() != 1 {
		t.Fatalf("sync=%d async=%d, want 1 and 1", syncHits.Load(), asyncHits.Load())
	}
}

// Under -race: concurrent subscribe/unsubscribe churn alongside concurrent
// publishers must neither race nor deadlock, and a subscriber that exists for
// the whole run receives every publish exactly once. Churned subscribers'
// counts are deliberately not asserted — they are racing with the publishes
// by construction.
func TestSubscribe_ConcurrentChurnKeepsAPinnedSubscriberWhole(t *testing.T) {
	bus := newBus()
	var pinned, churned atomic.Int32
	pin := mustSubscribe(t, bus, func(*evtA) { pinned.Add(1) })

	const publishers, perPublisher, churners = 4, 50, 4
	stop := make(chan struct{})
	var churnWG sync.WaitGroup
	for i := 0; i < churners; i++ {
		churnWG.Add(1)
		go func() {
			defer churnWG.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				sub, err := bus.Subscribe(func(*evtA) { churned.Add(1) })
				if err != nil {
					t.Errorf("Subscribe under churn: %v", err)
					return
				}
				sub.Unsubscribe()
			}
		}()
	}

	var pubWG sync.WaitGroup
	for p := 0; p < publishers; p++ {
		pubWG.Add(1)
		go func() {
			defer pubWG.Done()
			for i := 0; i < perPublisher; i++ {
				bus.Publish(&evtA{})
			}
		}()
	}
	published := make(chan struct{})
	go func() {
		pubWG.Wait()
		close(published)
	}()
	select {
	case <-published:
	case <-time.After(10 * time.Second):
		t.Fatal("publishers did not finish — deadlock under concurrent churn")
	}
	close(stop)
	churnWG.Wait()
	pin.Unsubscribe()

	if got := pinned.Load(); got != publishers*perPublisher {
		t.Fatalf("pinned subscriber received %d of %d publishes", got, publishers*perPublisher)
	}
}
