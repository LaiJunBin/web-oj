package handler

import (
	"errors"
	"testing"

	"judge-backend/internal/app/judge_domain/events/rejudgeapprovalrequested"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
)

// Claim/release/retry now lives in the shared claimed subscription (see
// claimed_subscription_test.go); these tests drive the handler body and its
// error/nil verdicts.

func newRejudgeApprovalHandler(t *testing.T) (*RejudgeApprovalRequestedHandler, *judgemocks.RejudgeRepository) {
	rejudgeRepo := judgemocks.NewRejudgeRepository(t)
	return &RejudgeApprovalRequestedHandler{
		rejudgeRepo: rejudgeRepo,
		logger:      sharedlogger.NewMockLogger(),
	}, rejudgeRepo
}

// A commit failure is retryable: the admin's approval must not be silently
// lost on a transient Mongo error.
func TestRejudgeApproval_CommitError_ReturnsError(t *testing.T) {
	h, rejudgeRepo := newRejudgeApprovalHandler(t)
	e := rejudgeapprovalrequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("CommitJudgmentsByRejudgeID", "rejudge-1", true).Return(false, errors.New("mongo unavailable")).Once()

	assert.Error(t, h.RejudgeApprovalRequested(e))
}

// TestRejudgeApproval_NotApplied_IsBenignSkip is a regression test: a lost
// race against a concurrent reject (or a second approval) must not be
// treated as a failure. CommitJudgmentsByRejudgeID's conditional guard
// returns applied=false with a nil error for exactly this case, and the
// handler must simply stop with nil — retrying cannot change a decision
// someone else already made.
func TestRejudgeApproval_NotApplied_IsBenignSkip(t *testing.T) {
	h, rejudgeRepo := newRejudgeApprovalHandler(t)
	e := rejudgeapprovalrequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("CommitJudgmentsByRejudgeID", "rejudge-1", true).Return(false, nil).Once()

	assert.NoError(t, h.RejudgeApprovalRequested(e))
}

func TestRejudgeApproval_Applied_Succeeds(t *testing.T) {
	h, rejudgeRepo := newRejudgeApprovalHandler(t)
	e := rejudgeapprovalrequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("CommitJudgmentsByRejudgeID", "rejudge-1", true).Return(true, nil).Once()

	assert.NoError(t, h.RejudgeApprovalRequested(e))
}

// TestRejudgeApproval_WiringPin_ClaimProtocolAndKey pins the registration
// (same template as TestCancelJudge_WiringPin_ClaimProtocolAndKey): the
// claimed protocol must be attempted with the event's own ID as the claim
// key; the denied claim keeps every other mock expectation-free.
func TestRejudgeApproval_WiringPin_ClaimProtocolAndKey(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	eventRepo := sharedmocks.NewEventRepository(t)
	eventRepo.On("TryClaimEvent", "evt-1").Return(false, nil).Once()

	RegisterRejudgeApprovalRequestedHandler(eb, eventRepo, judgemocks.NewRejudgeRepository(t), logger)

	eb.Publish(rejudgeapprovalrequested.New("evt-1", "rejudge-1"))
	eb.WaitAsync()
}
