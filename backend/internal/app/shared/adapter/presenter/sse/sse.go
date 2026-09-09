// Package sse renders an event-bus subscription as a Server-Sent Events body.
//
// It exists to fix a goroutine leak that every hand-rolled stream presenter
// shared: they subscribed with a synchronous callback that sent onto an
// UNBUFFERED channel, so a client that stopped reading (or merely read slowly)
// blocked the callback — which runs on the PUBLISHER's goroutine — until the
// request context was cancelled. Because several of these events are published
// from their own per-event goroutines, and the public scoreboard/time streams
// are unauthenticated, anyone could open a stream, stop reading, and pin an
// unbounded number of publisher goroutines until the process ran out of memory.
package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"

	sharedtypes "judge-backend/internal/app/shared/types"
)

// eventChannelBuffer bounds how many undelivered events one subscriber may hold
// before it is considered too slow. Small enough that a stalled client cannot
// accumulate unbounded memory, large enough to absorb an ordinary burst.
const eventChannelBuffer = 256

// Writer is the sink handed to callers for emitting SSE frames. It hides the
// buffering and framing so a presenter only decides WHAT to send.
type Writer struct {
	bw *bufio.Writer
}

// WriteComment writes an SSE comment line (": <text>"), which every client
// ignores by spec. Its purpose is bytes-on-the-wire: a stream whose writeInit
// sends nothing keeps the response headers sitting in the server's buffer
// until the FIRST event fires, so the client cannot even tell the stream
// opened. One comment up front flushes the headers and makes "connected" and
// "nothing has happened yet" distinguishable.
func (w *Writer) WriteComment(text string) error {
	if _, err := w.bw.WriteString(": " + text + "\n\n"); err != nil {
		return err
	}
	return w.bw.Flush()
}

// WriteData marshals v to JSON and writes it as one SSE `data:` frame, flushing
// immediately so the client sees each event as it happens. An error here is
// almost always the client having gone away; the caller should stop.
func (w *Writer) WriteData(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := w.bw.WriteString("data: " + string(b) + "\n\n"); err != nil {
		return err
	}
	return w.bw.Flush()
}

// Stream subscribes to events of type E and returns the read end of a pipe for
// the framework to copy to the client. The generic E must be the exact handler
// argument type the event bus keys its topic off (the same type the old
// callbacks used), e.g. *submissionupdated.Event.
//
// writeInit, if non-nil, emits an initial snapshot before the event loop.
// handle is called for each event: it should emit frames via w.WriteData and
// return nil (writing nothing is how you skip an event); any error — including
// a client-gone write error — ends the stream.
//
// Backpressure is the whole point: the subscription callback never blocks. If
// the buffer is full (client not keeping up) the stream is torn down instead of
// blocking the publisher, and a write already stalled inside the pipe is
// unblocked by closing the read end on teardown, so no goroutine outlives the
// client.
func Stream[E any](
	ctx context.Context,
	bus sharedtypes.EventBus,
	writeInit func(w *Writer) error,
	handle func(w *Writer, event E) error,
) io.Reader {
	streamCtx, cancel := context.WithCancel(ctx)
	events := make(chan E, eventChannelBuffer)

	callback := func(event E) {
		// Never blocks the publisher: enqueue if there is room, give up if the
		// stream is already ending, otherwise tear it down for being too slow.
		select {
		case events <- event:
		case <-streamCtx.Done():
		default:
			cancel()
		}
	}
	// The topic is derived from callback's parameter type E, so E must keep
	// implementing Topic(). The returned handle is what teardown uses: it
	// removes exactly this stream's subscription — an identity-based
	// unsubscribe once removed whichever same-type stream the bus had
	// registered first, freezing the earliest viewer whenever any later one
	// left.
	sub, err := bus.Subscribe(callback)
	if err != nil {
		// A stream that will never receive an event must not pretend to be
		// live: without this, the client held an open connection that stayed
		// silent forever. The eventbus already logged the failure (it is a
		// programming error — the event type lacks a Topic method); here we
		// hand back an immediately-EOF body so the client sees the stream end
		// and can surface an error instead of spinning. No pipe or goroutine
		// exists yet, so there is nothing else to clean up.
		cancel()
		return strings.NewReader("")
	}

	pr, pw := io.Pipe()

	go func() {
		<-streamCtx.Done()
		// Under the bus's snapshot semantics one event already in flight may
		// still reach callback after this returns; it lands in the
		// non-blocking select above and is dropped.
		sub.Unsubscribe()
		// Unblock a writer stalled mid-frame because the client stopped reading.
		_ = pr.CloseWithError(context.Canceled)
	}()

	go func() {
		defer cancel() // if the writer exits first (client gone), trigger cleanup
		defer pw.Close()
		w := &Writer{bw: bufio.NewWriter(pw)}

		if writeInit != nil {
			if err := writeInit(w); err != nil {
				return
			}
		}

		for {
			select {
			case <-streamCtx.Done():
				return
			case event := <-events:
				if err := handle(w, event); err != nil {
					return
				}
			}
		}
	}()

	return pr
}
