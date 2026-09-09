package handler

import (
	"judge-backend/internal/app/judge_domain/events/rejudgeapprovalrequested"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeApprovalRequestedHandler struct {
	rejudgeRepo repository.RejudgeRepository
	logger      sharedtypes.Logger
}

func RegisterRejudgeApprovalRequestedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	rejudgeRepo repository.RejudgeRepository,
	logger sharedtypes.Logger,
) {
	handler := &RejudgeApprovalRequestedHandler{
		rejudgeRepo: rejudgeRepo,
		logger:      logger,
	}

	// Claimed subscription: CommitJudgmentsByRejudgeID is a conditional
	// FINISHED→APPROVED transition (applied at most once), so a transient
	// failure retries instead of leaving an approval the admin already sent
	// permanently uncommitted.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.RejudgeApprovalRequested, sharedeventbus.ClaimedOptions[*rejudgeapprovalrequested.Event]{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[RejudgeApprovalRequestedHandler] Failed to subscribe: %v", err)
	}
	logger.Infof("[RejudgeApprovalRequestedHandler] Registered rejudge approval requested handler")
}

func (h *RejudgeApprovalRequestedHandler) RejudgeApprovalRequested(e *rejudgeapprovalrequested.Event) error {
	h.logger.Debugf("[RejudgeApprovalRequestedHandler] Processing rejudge approval requested event for rejudge %s", e.RejudgeID())

	applied, err := h.rejudgeRepo.CommitJudgmentsByRejudgeID(e.RejudgeID(), true)
	if err != nil {
		h.logger.Errorf("[RejudgeApprovalRequestedHandler] Failed to commit judgments for rejudge %s: %v", e.RejudgeID(), err)
		return err
	}
	if !applied {
		// Benign: someone already decided this rejudge; nothing to retry.
		h.logger.Debugf("[RejudgeApprovalRequestedHandler] Rejudge %s was no longer Finished (already approved/rejected concurrently), skipping", e.RejudgeID())
		return nil
	}
	h.logger.Infof("[RejudgeApprovalRequestedHandler] Successfully committed judgments for rejudge %s", e.RejudgeID())
	return nil
}
