package handler

import (
	"strings"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/events/judgmentcreated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	domainqueue "judge-backend/internal/app/judge_domain/judge/domain/queue"
	infraqueue "judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/stretchr/testify/mock"
)

// newResetRaceRig wires a real EventBus + real JudgeQueueMap and registers
// the handler through the REAL RegisterJudgmentCreatedHandler — literally the
// production wiring, claimed subscription included (the claim store is a mock
// granting the claim; retry-in-barrier semantics are covered by the eventbus
// layer's claimed_subscription tests). The priority strategy lookup is a
// controllable mock: its GetStrategyNameByContestProjectID call signals
// `started` and then blocks on `unblock`. That gives the test precise control
// over when the handler's goroutine actually reaches GetQueue().Push() —
// reproducing the moment a judgment_created handler is "in flight but hasn't
// pushed yet" that the BDD reset seam races against.
func newResetRaceRig(t *testing.T) (eventbus sharedtypes.EventBus, qp domainqueue.JudgeQueueProvider, publish func(), started, unblock chan struct{}) {
	logger := sharedlogger.NewMockLogger()
	eventbus = sharedeventbus.NewEventBus(logger)
	qp = infraqueue.NewJudgeQueueMap(logger)

	started = make(chan struct{})
	unblock = make(chan struct{})

	strategyQuery := judgemocks.NewPriorityStrategyQuery(t)
	strategyQuery.On("GetStrategyNameByContestProjectID", mock.Anything).
		Run(func(mock.Arguments) {
			close(started)
			<-unblock
		}).
		Return(prioritystrategies.NewDefault().Name(), nil).Once()

	strategyProvider := judgemocks.NewProvider(t)
	strategyProvider.On("GetStrategy", mock.Anything).Return(prioritystrategies.NewDefault(), nil).Once()

	maxDurationQuery := judgemocks.NewMaxJudgeDurationQuery(t)
	maxDurationQuery.On("GetMaxJudgeDurationMinutesByContestProjectID", mock.Anything).
		Return(nil, nil).Maybe()

	// The claim is granted (true): this rig needs the handler to RUN — the
	// race under test happens inside the handler body. The strict .Once()
	// doubles as the wiring pin: were the registration ever downgraded to a
	// raw SubscribeAsync, no claim would be attempted and this expectation
	// would fail the test.
	eventRepo := sharedmocks.NewEventRepository(t)
	eventRepo.On("TryClaimEvent", mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, "judgment_created-judgment-1-")
	})).Return(true, nil).Once()

	RegisterJudgmentCreatedHandler(
		eventbus, eventRepo,
		strategyQuery, maxDurationQuery, strategyProvider, qp, logger,
	)

	j := judgment.New(&judgment.JudgmentConfig{
		Project:   project.NewProject(&project.ProjectConfig{ID: 1}),
		JudgeType: judgetype.JUDGE,
		CreatedAt: time.Now(),
	})
	j.SetID("judgment-1")
	event := judgmentcreated.NewJudgmentCreatedEvent(j)

	return eventbus, qp, func() { eventbus.Publish(event) }, started, unblock
}

// TestJudgeSeam_ResetBeforeHandlerDrains_LeaksJudgmentPastReset reproduces
// the finding: if Reset() runs while a judgment_created handler is still in
// flight (published but not yet pushed), the handler pushes into the FRESH
// post-reset queue once it resumes — a judgment from the "old" scenario
// bleeding into the "new" one, invisible to Reset() because the push hadn't
// happened yet when Reset() ran.
func TestJudgeSeam_ResetBeforeHandlerDrains_LeaksJudgmentPastReset(t *testing.T) {
	_, qp, publish, started, unblock := newResetRaceRig(t)

	publish()
	<-started // handler is now blocked, mid-flight, has not pushed yet

	qp.Reset() // the old, buggy seam order: reset without draining first

	close(unblock) // let the handler resume; it pushes into the queue Reset() just replaced

	deadline := time.After(2 * time.Second)
	for {
		if j, _ := qp.TryPop(judgetype.JUDGE.String()); j != nil {
			return // leaked through, as the finding describes
		}
		select {
		case <-deadline:
			t.Fatal("expected the in-flight handler's judgment to leak into the post-reset queue")
		case <-time.After(time.Millisecond):
		}
	}
}

// TestJudgeSeam_DrainBeforeReset_NoLeak is the fix's counterpart: draining
// in-flight handlers (eventbus.WaitAsync, wrapped by waitAsyncWithTimeout in
// the real seam) before Reset() guarantees the handler's push lands in the
// PRE-reset queue, so Reset() correctly discards it.
func TestJudgeSeam_DrainBeforeReset_NoLeak(t *testing.T) {
	eventbus, qp, publish, started, unblock := newResetRaceRig(t)

	publish()
	<-started // handler is blocked, mid-flight, has not pushed yet

	seamDone := make(chan struct{})
	go func() {
		defer close(seamDone)
		eventbus.WaitAsync() // the fix: drain in-flight handlers first
		qp.Reset()
	}()

	// Give the seam goroutine a moment to actually be parked inside
	// WaitAsync (not required for correctness, just makes the ordering this
	// test is asserting on unambiguous) before letting the handler resume.
	time.Sleep(20 * time.Millisecond)
	close(unblock)

	select {
	case <-seamDone:
	case <-time.After(2 * time.Second):
		t.Fatal("seam goroutine (WaitAsync then Reset) did not complete")
	}

	got, err := qp.TryPop(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("TryPop failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected Reset (which ran only after the handler's push completed) to have discarded the judgment, got %s", got.ID())
	}
}
