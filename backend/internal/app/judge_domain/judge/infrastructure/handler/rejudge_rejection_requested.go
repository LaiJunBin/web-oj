package handler

import (
	"judge-backend/internal/app/judge_domain/events/rejudgerejectionrequested"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeRejectionRequestedHandler struct {
	rejudgeRepo repository.RejudgeRepository
	logger      sharedtypes.Logger
}

func RegisterRejudgeRejectionRequestedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	rejudgeRepo repository.RejudgeRepository,
	logger sharedtypes.Logger,
) {
	handler := &RejudgeRejectionRequestedHandler{
		rejudgeRepo: rejudgeRepo,
		logger:      logger,
	}

	// Claimed subscription: same conditional-commit shape as the
	// approval handler, retried for the same reason.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.RejudgeRejectionRequested, sharedeventbus.ClaimedOptions[*rejudgerejectionrequested.Event]{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[RejudgeRejectionRequestedHandler] Failed to subscribe: %v", err)
	}
	logger.Infof("[RejudgeRejectionRequestedHandler] Registered rejudge rejection requested handler")
}

func (h *RejudgeRejectionRequestedHandler) RejudgeRejectionRequested(e *rejudgerejectionrequested.Event) error {
	h.logger.Debugf("[RejudgeRejectionRequestedHandler] Processing rejudge rejection requested event for rejudge %s", e.RejudgeID())

	applied, err := h.rejudgeRepo.CommitJudgmentsByRejudgeID(e.RejudgeID(), false)
	if err != nil {
		h.logger.Errorf("[RejudgeRejectionRequestedHandler] Failed to commit judgments for rejudge %s: %v", e.RejudgeID(), err)
		return err
	}
	if !applied {
		// Benign: someone already decided this rejudge; nothing to retry.
		h.logger.Debugf("[RejudgeRejectionRequestedHandler] Rejudge %s was no longer Finished (already approved/rejected concurrently), skipping", e.RejudgeID())
		return nil
	}
	h.logger.Debugf("[RejudgeRejectionRequestedHandler] Successfully committed judgments for rejudge %s", e.RejudgeID())
	return nil
}
