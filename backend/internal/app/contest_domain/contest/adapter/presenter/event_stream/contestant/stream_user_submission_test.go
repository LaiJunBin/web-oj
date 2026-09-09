package espresenter_test

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	espresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/event_stream/contestant"
	streamusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_submission"
	"judge-backend/internal/app/contest_domain/events/submissionupdated"
	"judge-backend/internal/app/contest_domain/events/types"
	sharedtypes "judge-backend/internal/app/shared/types"

	"judge-backend/api"

	"github.com/stretchr/testify/assert"
)

// fakeBus captures the stream's subscribed callback so the test can push
// events into it directly, standing in for the judge pipeline.
type fakeBus struct {
	mu       sync.Mutex
	handlers []any
}

type fakeSubscription struct{}

func (s *fakeSubscription) Unsubscribe() {}

func (b *fakeBus) Publish(sharedtypes.Event) {}
func (b *fakeBus) Subscribe(handler any) (sharedtypes.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
	return &fakeSubscription{}, nil
}
func (b *fakeBus) SubscribeAsync(any, bool) error { return nil }
func (b *fakeBus) WaitAsync()                     {}

func (b *fakeBus) callback(t *testing.T) func(*submissionupdated.Event) {
	t.Helper()
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.handlers) != 1 {
		t.Fatalf("expected exactly one subscribed handler, got %d", len(b.handlers))
	}
	cb, ok := b.handlers[0].(func(*submissionupdated.Event))
	if !ok {
		t.Fatalf("subscribed handler has type %T", b.handlers[0])
	}
	return cb
}

// readFrame reads one SSE `data:` frame (up to its terminating blank line).
func readFrame(t *testing.T, br *bufio.Reader) string {
	t.Helper()
	var sb strings.Builder
	deadline := time.After(5 * time.Second)
	done := make(chan string, 1)
	go func() {
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				done <- sb.String()
				return
			}
			if line == "\n" {
				done <- sb.String()
				return
			}
			sb.WriteString(line)
		}
	}()
	select {
	case frame := <-done:
		return frame
	case <-deadline:
		t.Fatal("timed out waiting for an SSE frame")
		return ""
	}
}

func judgmentEvent(submissionID string, isRejudge bool, score float64) *submissionupdated.Event {
	return submissionupdated.New(10, types.JudgmentData{
		SubmissionID: submissionID,
		IsRejudge:    isRejudge,
		Status:       "Judging",
		Score:        score,
		GradingItems: []types.GradingItemData{
			{
				ID:     1,
				Score:  score,
				Status: "Judging",
				Testcases: []types.TestcaseData{
					{ID: 1, Status: "Passed", Stdout: "sample-visible-output", HasResult: true},
					{ID: 2, Status: "Failed", Stdout: "hidden-secret-output", Stderr: "hidden-secret-stderr", HasResult: true},
				},
			},
		},
	})
}

func fixedSampleSet(set map[int32]bool) func() map[int32]bool {
	return func() map[int32]bool { return set }
}

func streamBody(t *testing.T, ctx context.Context, bus *fakeBus, output *streamusersubmission.Output) io.Reader {
	t.Helper()
	p := espresenter.NewStreamUserSubmissionPresenter(bus)
	resp := p.Output(ctx, output)
	body, ok := resp.(api.StreamUserSubmission200TexteventStreamResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", resp)
	}
	return body.Body
}

// The stream must apply the same visibility rule as the REST detail: a
// testcase that is not one of the problem's samples is absent from the pushed
// frame entirely — id, status, stdout, stderr, all of it.
func TestStreamUserSubmission_HidesNonSampleTestcases(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := &fakeBus{}

	body := streamBody(t, ctx, bus, &streamusersubmission.Output{
		SubmissionID:             "sub-1",
		ResolveSampleTestcaseIDs: fixedSampleSet(map[int32]bool{1: true}),
	})
	br := bufio.NewReader(body)

	// The opening comment flushes the headers; without it the stream is
	// indistinguishable from a dead connection until the first judge update.
	assert.True(t, strings.HasPrefix(readFrame(t, br), ": "), "expected the stream to open with an SSE comment")

	bus.callback(t)(judgmentEvent("sub-1", false, 42))

	frame := readFrame(t, br)
	assert.Contains(t, frame, "sample-visible-output")
	assert.Contains(t, frame, `"score":42`)
	assert.NotContains(t, frame, "hidden-secret-output")
	assert.NotContains(t, frame, "hidden-secret-stderr")
}

// Pins of the existing skip rules, so the filter cannot loosen them.
func TestStreamUserSubmission_SkipsRejudgeAndForeignEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := &fakeBus{}

	body := streamBody(t, ctx, bus, &streamusersubmission.Output{
		SubmissionID:             "sub-1",
		ResolveSampleTestcaseIDs: fixedSampleSet(map[int32]bool{1: true}),
	})
	br := bufio.NewReader(body)
	assert.True(t, strings.HasPrefix(readFrame(t, br), ": "), "expected the stream to open with an SSE comment")

	cb := bus.callback(t)
	cb(judgmentEvent("sub-1", true, 1))      // rejudge: must write nothing
	cb(judgmentEvent("sub-other", false, 2)) // another submission: nothing
	cb(judgmentEvent("sub-1", false, 3))     // this one produces the frame

	frame := readFrame(t, br)
	assert.Contains(t, frame, `"score":3`)
	assert.NotContains(t, frame, `"score":1`)
	assert.NotContains(t, frame, `"score":2`)
}

// The resolver is consulted per frame, so a visibility change lands on the
// very next event once the resolver reflects it — an open stream no longer
// freezes the set it saw at open.
func TestStreamUserSubmission_FollowsResolverChanges(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := &fakeBus{}

	current := map[int32]bool{1: true}
	body := streamBody(t, ctx, bus, &streamusersubmission.Output{
		SubmissionID:             "sub-1",
		ResolveSampleTestcaseIDs: func() map[int32]bool { return current },
	})
	br := bufio.NewReader(body)
	assert.True(t, strings.HasPrefix(readFrame(t, br), ": "), "expected the stream to open with an SSE comment")

	cb := bus.callback(t)
	cb(judgmentEvent("sub-1", false, 1))
	first := readFrame(t, br)
	assert.Contains(t, first, "sample-visible-output")
	assert.NotContains(t, first, "hidden-secret-output")

	// Testcase 1 goes hidden, testcase 2 becomes a sample.
	current = map[int32]bool{2: true}
	cb(judgmentEvent("sub-1", false, 2))
	second := readFrame(t, br)
	assert.Contains(t, second, "hidden-secret-output", "the newly visible testcase must stream")
	assert.NotContains(t, second, "sample-visible-output", "the newly hidden testcase must stop streaming")
}

// An ERROR is pushed live; its reason must ride on the frame or the
// contestant has to reload to learn there is nothing to fix on their side.
func TestStreamUserSubmission_CarriesReasonOnErrorFrames(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := &fakeBus{}

	body := streamBody(t, ctx, bus, &streamusersubmission.Output{
		SubmissionID:             "sub-1",
		ResolveSampleTestcaseIDs: fixedSampleSet(map[int32]bool{1: true}),
	})
	br := bufio.NewReader(body)
	assert.True(t, strings.HasPrefix(readFrame(t, br), ": "))

	reason := "評測前置步驟失敗，請聯絡現場工作人員"
	bus.callback(t)(submissionupdated.New(10, types.JudgmentData{
		SubmissionID: "sub-1", Status: "Error", Score: 0, Reason: &reason,
	}))
	frame := readFrame(t, br)
	assert.Contains(t, frame, `"status":"Error"`)
	assert.Contains(t, frame, reason)

	// And a frame with nothing to explain omits the field entirely.
	bus.callback(t)(judgmentEvent("sub-1", false, 5))
	assert.NotContains(t, readFrame(t, br), `"reason"`)
}

// A staff note is pushed live on the same frame as everything else, and the
// two explanations are independent: an Error frame can carry both the
// system reason and the operator's note.
func TestStreamUserSubmission_CarriesStaffNoteAlongsideReason(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := &fakeBus{}

	body := streamBody(t, ctx, bus, &streamusersubmission.Output{
		SubmissionID:             "sub-1",
		ResolveSampleTestcaseIDs: fixedSampleSet(map[int32]bool{1: true}),
	})
	br := bufio.NewReader(body)
	assert.True(t, strings.HasPrefix(readFrame(t, br), ": "))

	reason := "評測前置步驟失敗，請聯絡現場工作人員"
	note := "已通知助教，請稍候"
	bus.callback(t)(submissionupdated.New(10, types.JudgmentData{
		SubmissionID: "sub-1", Status: "Error", Reason: &reason, StaffNote: &note,
	}))
	frame := readFrame(t, br)
	assert.Contains(t, frame, `"reason":"`+reason+`"`)
	assert.Contains(t, frame, `"staffNote":"`+note+`"`)

	// A note on a healthy judgment stands alone — no reason is invented.
	bus.callback(t)(submissionupdated.New(10, types.JudgmentData{
		SubmissionID: "sub-1", Status: "Judged", Score: 100, StaffNote: &note,
	}))
	frame = readFrame(t, br)
	assert.Contains(t, frame, `"staffNote":"`+note+`"`)
	assert.NotContains(t, frame, `"reason"`)
}

// Clearing a note publishes a frame WITHOUT the field. That absence is the
// signal — the client must drop the note it was showing, not keep the last
// value the way it may for reason.
func TestStreamUserSubmission_OmitsStaffNoteOnceCleared(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := &fakeBus{}

	body := streamBody(t, ctx, bus, &streamusersubmission.Output{
		SubmissionID:             "sub-1",
		ResolveSampleTestcaseIDs: fixedSampleSet(map[int32]bool{1: true}),
	})
	br := bufio.NewReader(body)
	assert.True(t, strings.HasPrefix(readFrame(t, br), ": "))

	note := "暫時的附註"
	cb := bus.callback(t)
	cb(submissionupdated.New(10, types.JudgmentData{SubmissionID: "sub-1", Status: "Judged", StaffNote: &note}))
	assert.Contains(t, readFrame(t, br), `"staffNote"`)

	cb(submissionupdated.New(10, types.JudgmentData{SubmissionID: "sub-1", Status: "Judged", StaffNote: nil}))
	assert.NotContains(t, readFrame(t, br), `"staffNote"`)
}
