package eventbus

import (
	"fmt"
	sharedtypes "judge-backend/internal/app/shared/types"
	"reflect"
	"sync"

	"github.com/asaskevich/EventBus"
)

// Eventbus wraps asaskevich/EventBus.
//
// Synchronous subscriptions are NOT handed to the library one by one. The
// library identifies a handler for removal by reflect Type() + Pointer(), and
// Pointer() on a func is its CODE pointer — every closure stamped out of one
// func literal (sse.Stream's callback, for instance) shares it — so an
// identity-based Unsubscribe removed whichever same-type handler had been
// registered FIRST: one viewer leaving a stream silently froze the earliest
// viewer of that event type. Instead the wrapper registers ONE fan-out
// handler per topic with the library and keeps its own subscriber list; each
// subscriber is owned by the Subscription handle that Subscribe returns, and
// only that handle can remove it.
//
// Async subscriptions (SubscribeAsync, and the claimed/retried helpers built
// on it) are untouched: they register with the library directly and are never
// unsubscribed.
type Eventbus struct {
	eventBus EventBus.Bus
	logger   sharedtypes.Logger

	// regMu guards registered — the topics whose fan-out handler already
	// exists in the library. It is held while calling into the library (which
	// takes its own lock), and never while subsMu is held.
	regMu      sync.Mutex
	registered map[string]bool

	// subsMu guards subs. The fan-out takes it briefly, for reading, to
	// snapshot a topic's list while the library's lock is held; nothing that
	// holds subsMu ever calls into the library. Lock order is therefore
	// bus.lock → subsMu (publish) and regMu → bus.lock (first subscribe) with
	// no path back — no cycle.
	subsMu sync.RWMutex
	subs   map[string][]*syncSub
}

type syncSub struct {
	topic string
	fn    reflect.Value
}

func NewEventBus(logger sharedtypes.Logger) sharedtypes.EventBus {
	return &Eventbus{
		eventBus:   EventBus.New(),
		logger:     logger,
		registered: map[string]bool{},
		subs:       map[string][]*syncSub{},
	}
}

func (e *Eventbus) Publish(event sharedtypes.Event) {
	e.logger.Debugf("Publishing event: %s", event.Topic())
	e.eventBus.Publish(event.Topic(), event)
}

// Subscribe registers a synchronous handler and returns the handle that
// removes it. See sharedtypes.EventBus for the handler contract.
func (e *Eventbus) Subscribe(handler interface{}) (sharedtypes.Subscription, error) {
	// Validated here, ahead of getHandlerTopic: that helper indexes In(0)
	// unconditionally and would panic on a zero-argument func or a nil.
	t := reflect.TypeOf(handler)
	if t == nil || t.Kind() != reflect.Func || t.NumIn() != 1 {
		err := fmt.Errorf("handler must be a func with exactly one event parameter, got %v", t)
		e.logger.Errorf("Failed to subscribe handler: %v", err)
		return nil, err
	}
	topic, err := getHandlerTopic(handler)
	if err != nil {
		e.logger.Errorf("Failed to subscribe handler: %v", err)
		return nil, err
	}
	if err := e.ensureFanOut(topic); err != nil {
		e.logger.Errorf("Failed to subscribe handler: %v", err)
		return nil, err
	}

	s := &syncSub{topic: topic, fn: reflect.ValueOf(handler)}
	e.subsMu.Lock()
	e.subs[topic] = append(e.subs[topic], s)
	e.subsMu.Unlock()

	e.logger.Debugf("Subscribing to topic: %s", topic)
	return &subscription{bus: e, sub: s}, nil
}

// ensureFanOut registers the topic's single library-side handler on first
// use. It is never removed (the count is bounded by the number of event
// types); what comes and goes is the subscriber list behind it. regMu is held
// across the library call so concurrent first subscribers of one topic
// serialize here, and every Subscribe returns only once the fan-out exists —
// a Publish after any Subscribe has returned reaches that subscriber.
func (e *Eventbus) ensureFanOut(topic string) error {
	e.regMu.Lock()
	defer e.regMu.Unlock()
	if e.registered[topic] {
		return nil
	}
	if err := e.eventBus.Subscribe(topic, e.fanOutFor(topic)); err != nil {
		return err
	}
	e.registered[topic] = true
	return nil
}

// fanOutFor builds the one library handler for a topic. The library calls it
// on the publisher's goroutine with its lock held — exactly where a
// synchronous handler used to run — passing the concrete event, which
// reflect then hands to each subscriber's typed parameter. The list is
// snapshotted and released BEFORE any subscriber runs, so a subscriber may
// unsubscribe (itself or anyone) from inside its callback without deadlock;
// the price is that a subscriber can see at most one delivery that was
// snapshotted before its Unsubscribe returned.
func (e *Eventbus) fanOutFor(topic string) func(event interface{}) {
	return func(event interface{}) {
		e.subsMu.RLock()
		current := e.subs[topic]
		if len(current) == 0 {
			e.subsMu.RUnlock()
			return
		}
		snapshot := make([]*syncSub, len(current))
		copy(snapshot, current)
		e.subsMu.RUnlock()

		args := []reflect.Value{reflect.ValueOf(event)}
		for _, s := range snapshot {
			s.fn.Call(args)
		}
	}
}

// subscription is the handle behind sharedtypes.Subscription.
type subscription struct {
	bus *Eventbus
	sub *syncSub
}

// Unsubscribe removes exactly this subscriber. Copy-on-write, never an
// in-place shift: a fan-out may be walking a snapshot of the current slice.
// Idempotent, and it never touches the library — which is the whole point.
func (s *subscription) Unsubscribe() {
	s.bus.subsMu.Lock()
	defer s.bus.subsMu.Unlock()
	current := s.bus.subs[s.sub.topic]
	replacement := make([]*syncSub, 0, len(current))
	for _, c := range current {
		if c != s.sub {
			replacement = append(replacement, c)
		}
	}
	if len(replacement) == len(current) {
		return // already gone
	}
	s.bus.subs[s.sub.topic] = replacement
	s.bus.logger.Debugf("Unsubscribing from topic: %s", s.sub.topic)
}

// SubscribeAsync hands the handler to the library directly. These handlers are
// never unsubscribed, and identity-based removal must never be added here:
// the library would match by code pointer, so the closure every
// SubscribeClaimedAsync/SubscribeRetriedAsync call wraps its handler in (one
// func literal) and the method values the handlers are registered with (all
// method values share a pointer) are indistinguishable to it. If async
// unsubscription is ever needed, route it through a wrapper-side registry as
// Subscribe does.
func (e *Eventbus) SubscribeAsync(handler interface{}, transactional bool) error {
	topic, err := getHandlerTopic(handler)
	if err != nil {
		e.logger.Errorf("Failed to subscribe handler asynchronously: %v", err)
		return err
	}
	e.logger.Debugf("Subscribing asynchronously to topic: %s", topic)
	return e.eventBus.SubscribeAsync(topic, handler, transactional)
}

func getHandlerTopic(handler interface{}) (string, error) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		return "", fmt.Errorf("%s is not of type reflect.Func", reflect.TypeOf(handler).Kind())
	}

	event := reflect.TypeOf(handler).In(0)
	topicFn, ok := event.MethodByName("Topic")
	if !ok {
		return "", fmt.Errorf("event does not have a Topic method")
	}

	topic := topicFn.Func.Call([]reflect.Value{reflect.New(event).Elem()})[0].String()
	return topic, nil
}

func (e *Eventbus) WaitAsync() {
	e.eventBus.WaitAsync()
}
