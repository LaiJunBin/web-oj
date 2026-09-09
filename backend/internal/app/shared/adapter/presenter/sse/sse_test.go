package sse

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	sharedtypes "judge-backend/internal/app/shared/types"
)

// fakeBus is a minimal EventBus that captures every subscribed handler, in
// order, so a test can invoke one directly — and records, by index, which
// subscription each returned handle removed. Holding several handlers is the
// point: a single-handler fake could never show that a stream tears down ITS
// OWN subscription rather than a neighbour's.
type fakeBus struct {
	mu           sync.Mutex
	handlers     []any
	removed      []int
	subscribeErr error
}

type fakeSubscription struct {
	bus *fakeBus
	idx int
}

func (s *fakeSubscription) Unsubscribe() {
	s.bus.mu.Lock()
	defer s.bus.mu.Unlock()
	s.bus.removed = append(s.bus.removed, s.idx)
}

func (b *fakeBus) Publish(sharedtypes.Event) {}
func (b *fakeBus) Subscribe(handler any) (sharedtypes.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscribeErr != nil {
		return nil, b.subscribeErr
	}
	b.handlers = append(b.handlers, handler)
	return &fakeSubscription{bus: b, idx: len(b.handlers) - 1}, nil
}
func (b *fakeBus) SubscribeAsync(any, bool) error { return nil }
func (b *fakeBus) WaitAsync()                     {}

// callback returns the one subscribed handler of a single-stream test.
func (b *fakeBus) callback(t *testing.T) func(*evt) {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.handlers) != 1 {
		t.Fatalf("expected exactly one subscribed handler, got %d", len(b.handlers))
	}
	cb, ok := b.handlers[0].(func(*evt))
	if !ok {
		t.Fatalf("handler has wrong type: %T", b.handlers[0])
	}
	return cb
}

// didUnsubscribe reports whether any subscription's handle has been used.
func (b *fakeBus) didUnsubscribe() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.removed) > 0
}

func (b *fakeBus) removedIndexes() []int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]int(nil), b.removed...)
}

type evt struct {
	N int `json:"n"`
}

func (e *evt) Topic() string { return "test:evt" }

// TestStream_SubscribeFailureEndsTheStreamImmediately: a subscription that
// cannot be established must not hand the client a silent, forever-open body.
// The stream ends at once (EOF), and since nothing was subscribed there is
// nothing to unsubscribe.
func TestStream_SubscribeFailureEndsTheStreamImmediately(t *testing.T) {
	bus := &fakeBus{subscribeErr: errors.New("no Topic method")}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	body := Stream(ctx, bus, nil, func(w *Writer, e *evt) error {
		return w.WriteData(e)
	})

	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(body)
		done <- result{data, err}
	}()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("expected clean EOF, got error: %v", got.err)
		}
		if len(got.data) != 0 {
			t.Fatalf("expected empty body, got %q", got.data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("body never reached EOF — the stream pretended to be live")
	}

	if bus.didUnsubscribe() {
		t.Fatal("nothing was subscribed, so nothing should be unsubscribed")
	}
}

// TestStream_DeliversToReadingClient is the happy path: an event handled by the
// stream reaches a reading client as a `data:` frame.
func TestStream_DeliversToReadingClient(t *testing.T) {
	bus := &fakeBus{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	body := Stream(ctx, bus, nil, func(w *Writer, e *evt) error {
		return w.WriteData(e)
	})

	cb := bus.callback(t)
	cb(&evt{N: 7})

	line := make(chan string, 1)
	go func() {
		r := bufio.NewReader(body)
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
		if !strings.Contains(got, `"n":7`) {
			t.Fatalf("expected frame with n:7, got %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SSE frame")
	}
}

// TestStream_StalledClientNeverBlocksPublisher is the regression test for the
// leak: with a client that never reads, flooding the subscription must not
// block the publisher's callback, and the stream must tear itself down
// (unsubscribe) rather than pinning a goroutine forever.
func TestStream_StalledClientNeverBlocksPublisher(t *testing.T) {
	bus := &fakeBus{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The client (body reader) is deliberately never read from, so the writer
	// goroutine stalls on the pipe after the first event.
	_ = Stream(ctx, bus, nil, func(w *Writer, e *evt) error {
		return w.WriteData(e)
	})
	cb := bus.callback(t)

	// Publish far more than the buffer can hold. Every call must return quickly;
	// if any blocked, this loop would not finish within the timeout.
	done := make(chan struct{})
	go func() {
		for i := 0; i < eventChannelBuffer*4; i++ {
			cb(&evt{N: i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("publisher callback blocked — the leak this fix prevents")
	}

	// Overflow must have torn the stream down and unsubscribed.
	deadline := time.Now().Add(2 * time.Second)
	for !bus.didUnsubscribe() {
		if time.Now().After(deadline) {
			t.Fatal("stream did not unsubscribe after overflow teardown")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestStream_ClientDisconnectUnsubscribes verifies the ordinary path: cancelling
// the request context unsubscribes and ends the stream.
func TestStream_ClientDisconnectUnsubscribes(t *testing.T) {
	bus := &fakeBus{}
	ctx, cancel := context.WithCancel(context.Background())

	_ = Stream(ctx, bus, nil, func(w *Writer, e *evt) error { return nil })
	_ = bus.callback(t) // ensure subscribed

	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for !bus.didUnsubscribe() {
		if time.Now().After(deadline) {
			t.Fatal("stream did not unsubscribe on context cancel")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestStream_TeardownUnsubscribesItsOwnHandle: with two streams on one bus,
// tearing the SECOND down must use the second subscription's handle — index
// 1 — and leave the first alone. The fake records indexes, so this pins "a
// stream ends only its own subscription" independently of the real bus; the
// real-bus counterpart is TestStream_TwoViewers_TearingDownOneLeavesTheOtherLive.
func TestStream_TeardownUnsubscribesItsOwnHandle(t *testing.T) {
	bus := &fakeBus{}
	handle := func(w *Writer, e *evt) error { return nil }

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	_ = Stream(ctx1, bus, nil, handle)

	ctx2, cancel2 := context.WithCancel(context.Background())
	_ = Stream(ctx2, bus, nil, handle)

	cancel2()

	deadline := time.Now().Add(2 * time.Second)
	for !bus.didUnsubscribe() {
		if time.Now().After(deadline) {
			t.Fatal("second stream did not unsubscribe on context cancel")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := bus.removedIndexes(); len(got) != 1 || got[0] != 1 {
		t.Fatalf("expected exactly the second subscription (index 1) to be removed, got %v", got)
	}
}
