package handler

import (
	"errors"
	"testing"

	"judge-backend/internal/app/judge_domain/events/rejudgerejectionrequested"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
)

// Claim/release/retry now lives in the shared claimed subscription (see
// claimed_subscription_test.go); these tests drive the handler body and its
// error/nil verdicts.

func newRejudgeRejectionHandler(t *testing.T) (*RejudgeRejectionRequestedHandler, *judgemocks.RejudgeRepository) {
	rejudgeRepo := judgemocks.NewRejudgeRepository(t)
	return &RejudgeRejectionRequestedHandler{
		rejudgeRepo: rejudgeRepo,
		logger:      sharedlogger.NewMockLogger(),
	}, rejudgeRepo
}

func TestRejudgeRejection_CommitError_ReturnsError(t *testing.T) {
	h, rejudgeRepo := newRejudgeRejectionHandler(t)
	e := rejudgerejectionrequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("CommitJudgmentsByRejudgeID", "rejudge-1", false).Return(false, errors.New("mongo unavailable")).Once()

	assert.Error(t, h.RejudgeRejectionRequested(e))
}

func TestRejudgeRejection_NotApplied_IsBenignSkip(t *testing.T) {
	h, rejudgeRepo := newRejudgeRejectionHandler(t)
	e := rejudgerejectionrequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("CommitJudgmentsByRejudgeID", "rejudge-1", false).Return(false, nil).Once()

	assert.NoError(t, h.RejudgeRejectionRequested(e))
}

func TestRejudgeRejection_Applied_Succeeds(t *testing.T) {
	h, rejudgeRepo := newRejudgeRejectionHandler(t)
	e := rejudgerejectionrequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("CommitJudgmentsByRejudgeID", "rejudge-1", false).Return(true, nil).Once()

	assert.NoError(t, h.RejudgeRejectionRequested(e))
}

// TestRejudgeRejection_WiringPin_ClaimProtocolAndKey pins the registration
// (same template as TestCancelJudge_WiringPin_ClaimProtocolAndKey): the
// claimed protocol must be attempted with the event's own ID as the claim
// key; the denied claim keeps every other mock expectation-free.
func TestRejudgeRejection_WiringPin_ClaimProtocolAndKey(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	eventRepo := sharedmocks.NewEventRepository(t)
	eventRepo.On("TryClaimEvent", "evt-1").Return(false, nil).Once()

	RegisterRejudgeRejectionRequestedHandler(eb, eventRepo, judgemocks.NewRejudgeRepository(t), logger)

	eb.Publish(rejudgerejectionrequested.New("evt-1", "rejudge-1"))
	eb.WaitAsync()
}
