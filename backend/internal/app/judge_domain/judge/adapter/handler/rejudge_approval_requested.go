package handler

import (
	"judge-backend/internal/app/judge_domain/events/rejudgeapprovalrequested"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeApprovalRequestedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterRejudgeApprovalRequestedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &RejudgeApprovalRequestedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[RejudgeApprovalRequestedHandler] Registered")
}

func (h *RejudgeApprovalRequestedHandler) Handle(e *sharedevent.RejudgeApprovalRequested) {
	h.logger.Debugf("[RejudgeApprovalRequestedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(rejudgeapprovalrequested.New(e.ID(), e.RejudgeID()))
}
