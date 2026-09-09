package handler

import (
	"testing"
	"time"

	adapterhandler "judge-backend/internal/app/judge_domain/judge/adapter/handler"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestCancelRejudge_CrossLayer_RepeatedCancelHealsStragglers wires the REAL
// event bus through both layers — the shared cancel-rejudge event, the
// adapter that translates it into the judge-domain event, and the
// infrastructure handler — and drives the exact recovery path the contest
// usecase's idempotent-retry carve-out exists for: the parent is ALREADY
// CANCELED (a previous attempt died partway through its members), the user
// requests cancellation again, and the straggler member must still be
// cancelled. Each request publishes a fresh event ID, so the claim-based
// dedup must not block the retry either.
func TestCancelRejudge_CrossLayer_RepeatedCancelHealsStragglers(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)

	eventRepo := sharedmocks.NewEventRepository(t)
	rejudgeRepo := judgemocks.NewRejudgeRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)

	adapterhandler.RegisterCancelRejudgeRequestedHandler(eb, logger)
	RegisterCancelRejudgeRequestedHandler(eb, eventRepo, rejudgeRepo, judgmentRepo, logger)

	parent := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-xl",
		Status:    rejudge.CANCELED,
		Judgments: []string{"judgment-straggler"},
	})

	// The retry carries a fresh UUID, minted inside the constructor — so the
	// event is built first and the claim key pinned to ITS exact ID. That
	// pins more than key hygiene: the ACL adapter must hand the shared
	// event's own ID down to the domain event, because a claim keyed on
	// anything stable (say the rejudge ID) would dedupe away exactly the
	// repeated-cancel healing this test protects.
	evt := sharedevent.NewCancelRejudgeRequested("rejudge-xl")
	eventRepo.On("TryClaimEvent", evt.(*sharedevent.CancelRejudgeRequested).ID()).Return(true, nil).Once()
	rejudgeRepo.On("FindByID", "rejudge-xl").Return(parent, nil).Twice()
	rejudgeRepo.On("TransitionStatus", "rejudge-xl", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(false, nil).Once()

	healed := make(chan struct{})
	judgmentRepo.On("TryCancel", "judgment-straggler", "", mock.AnythingOfType("time.Time")).
		Return(cancelledMember("judgment-straggler"), nil).
		Run(func(mock.Arguments) { close(healed) }).Once()

	// This is exactly what the contest-side RejudgeService publishes on a
	// repeated cancel request.
	eb.Publish(evt)

	select {
	case <-healed:
	case <-time.After(5 * time.Second):
		t.Fatal("the repeated cancel request never reached the member reconciliation across the event layers")
	}
	// Let the handler's trailing publishes (judge_cancelled / judgeupdated,
	// no subscribers here) drain before mock assertions run.
	eb.WaitAsync()
	require.True(t, true)
}
