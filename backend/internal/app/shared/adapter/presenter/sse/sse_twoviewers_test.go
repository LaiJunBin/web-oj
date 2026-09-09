package sse

import (
	"bufio"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
)

// signallingBus decorates a REAL eventbus and records every subscription it
// hands out, each with its own `done` channel closed once that stream's
// teardown has actually unsubscribed — the deterministic barrier this test
// needs. Draining the torn-down stream's pipe is NOT a barrier: its EOF comes
// from the writer goroutine's pw.Close, which does not order against the
// teardown goroutine's unsubscribe.
type signallingBus struct {
	sharedtypes.EventBus
	mu   sync.Mutex
	subs []*signallingSubscription
}

func (b *signallingBus) Subscribe(handler interface{}) (sharedtypes.Subscription, error) {
	sub, err := b.EventBus.Subscribe(handler)
	if err != nil {
		return nil, err
	}
	s := &signallingSubscription{inner: sub, done: make(chan struct{})}
	b.mu.Lock()
	b.subs = append(b.subs, s)
	b.mu.Unlock()
	return s, nil
}

func (b *signallingBus) subscription(t *testing.T, i int) *signallingSubscription {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()
	if i >= len(b.subs) {
		t.Fatalf("only %d subscriptions were made, wanted index %d", len(b.subs), i)
	}
	return b.subs[i]
}

// signallingSubscription signals only AFTER the real Unsubscribe returned,
// so a publish issued after the signal cannot find the subscriber.
type signallingSubscription struct {
	inner sharedtypes.Subscription
	done  chan struct{}
	once  sync.Once
}

func (s *signallingSubscription) Unsubscribe() {
	s.inner.Unsubscribe()
	s.once.Do(func() { close(s.done) })
}

// TestStream_TwoViewers_TearingDownOneLeavesTheOtherLive is the regression
// test for the wrong-subscriber-removed bug: two viewers of the same event
// type each hold a Stream; the SECOND one disconnects; an event published
// afterwards must still reach the FIRST. The underlying library identifies
// handlers by reflect code pointer, and every Stream[E] callback comes from
// one func literal, so an identity-based unsubscribe removed the earliest
// same-type viewer — the scoreboard big screen — instead of the one leaving.
func TestStream_TwoViewers_TearingDownOneLeavesTheOtherLive(t *testing.T) {
	bus := &signallingBus{EventBus: sharedeventbus.NewEventBus(sharedlogger.NewMockLogger())}
	handle := func(w *Writer, e *evt) error { return w.WriteData(e) }

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	first := Stream(ctx1, bus, nil, handle)

	ctx2, cancel2 := context.WithCancel(context.Background())
	_ = Stream(ctx2, bus, nil, handle)

	cancel2()
	select {
	case <-bus.subscription(t, 1).done:
	case <-time.After(2 * time.Second):
		t.Fatal("the second stream's teardown never unsubscribed")
	}

	bus.Publish(&evt{N: 42})

	line := make(chan string, 1)
	go func() {
		r := bufio.NewReader(first)
		for {
			s, err := r.ReadString('\n')
			if strings.HasPrefix(s, "data:") {
				line <- strings.TrimSpace(s)
				return
			}
			if err != nil {
				line <- "ERR:" + err.Error()
				return
			}
		}
	}()
	select {
	case got := <-line:
		if !strings.Contains(got, `"n":42`) {
			t.Fatalf("first viewer got %q, want the n:42 frame", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the first viewer received nothing after the second viewer left — its subscription was the one removed")
	}
}
