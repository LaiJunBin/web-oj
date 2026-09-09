package handler

import (
	"errors"
	"testing"

	"judge-backend/internal/app/judge_domain/events/canceljudgerequested"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// The claim/release/retry protocol now lives in the shared claimed
// subscription (see claimed_subscription_test.go); these tests exercise the
// handler body directly — its error/nil verdicts are what drive that
// protocol.

func newCancelJudgeHandler(t *testing.T) (*CancelJudgeRequested, *sharedmocks.EventBus, *judgemocks.JudgmentRepository) {
	eventbus := sharedmocks.NewEventBus(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	return &CancelJudgeRequested{
		eventbus:     eventbus,
		judgmentRepo: judgmentRepo,
		logger:       sharedlogger.NewMockLogger(),
	}, eventbus, judgmentRepo
}

// TryCancel failing is the retryable case: the handler must return the error
// (so the claimed subscription releases the claim and retries) and publish
// nothing.
func TestCancelJudge_TryCancelError_ReturnsErrorNoPublish(t *testing.T) {
	h, _, judgmentRepo := newCancelJudgeHandler(t)
	e := canceljudgerequested.New("evt-1", "judgment-1", "user requested")

	judgmentRepo.On("TryCancel", "judgment-1", "user requested", mock.AnythingOfType("time.Time")).Return(nil, errors.New("mongo unavailable")).Once()
	// No Publish expectation: an unexpected call would panic.

	assert.Error(t, h.CancelJudge(e))
}

// TestCancelJudge_AlreadyTerminal_NoPublish is a regression test for the
// same race TryCancel's doc comment describes: the worker can independently
// move this judgment ID to a terminal status (Judged/Rejudged/Error) between
// the event being published and this handler processing it. TryCancel's
// filter only matches Queued/Judging, so it reports applied=false instead of
// touching the document — the handler stops there without publishing
// judge_cancelled, and returns nil: a retry cannot change a terminal status.
func TestCancelJudge_AlreadyTerminal_NilNoPublish(t *testing.T) {
	h, _, judgmentRepo := newCancelJudgeHandler(t)
	e := canceljudgerequested.New("evt-1", "judgment-1", "user requested")

	judgmentRepo.On("TryCancel", "judgment-1", "user requested", mock.AnythingOfType("time.Time")).Return(nil, nil).Once()
	// No Publish expectation: an unexpected call would panic.

	assert.NoError(t, h.CancelJudge(e))
}

// TestCancelJudge_Applied_PublishesJudgeCancelledAndJudgeUpdated: having won
// the Cancelled transition, the handler owns making it visible — for a
// still-QUEUED judgment nobody else ever publishes anything (no worker is
// executing it, and its eventual TryStart just fails silently), so without
// the judgeupdated publish here the rejudge rollup never sees this member
// reach a terminal status and the contest stream never shows the
// cancellation.
func TestCancelJudge_Applied_PublishesJudgeCancelledAndJudgeUpdated(t *testing.T) {
	h, eventbus, judgmentRepo := newCancelJudgeHandler(t)
	e := canceljudgerequested.New("evt-1", "judgment-1", "user requested")

	// The entity comes from TryCancel's own atomic command — no FindByID
	// expectation is registered, so any re-read would panic.
	judgmentRepo.On("TryCancel", "judgment-1", "user requested", mock.AnythingOfType("time.Time")).Return(cancelledMember("judgment-1"), nil).Once()
	expectJudgeCancelledPublish(eventbus, "judgment-1")
	expectJudgeUpdatedCancelledPublish(eventbus, "judgment-1")

	assert.NoError(t, h.CancelJudge(e))
}

// TestCancelJudge_WiringPin_ClaimProtocolAndKey pins the REGISTRATION, not
// the handler: RegisterCancelJudgeRequestedHandler must subscribe through the
// claimed protocol with the event's own ID as the claim key. Publishing
// on a real bus with the claim DENIED proves the wiring without running the
// handler — every other dependency is a bare strict mock — so a registration
// downgraded to a raw SubscribeAsync (which never calls TryClaimEvent) or a
// changed claim key fails this test. The handler's behavior is covered by
// the tests above; claim/release/retry semantics by the eventbus-layer
// claimed_subscription tests.
func TestCancelJudge_WiringPin_ClaimProtocolAndKey(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	eventRepo := sharedmocks.NewEventRepository(t)
	eventRepo.On("TryClaimEvent", "evt-1").Return(false, nil).Once()

	RegisterCancelJudgeRequestedHandler(eb, eventRepo, judgemocks.NewJudgmentRepository(t), logger)

	eb.Publish(canceljudgerequested.New("evt-1", "judgment-1", "user requested"))
	eb.WaitAsync()
}
