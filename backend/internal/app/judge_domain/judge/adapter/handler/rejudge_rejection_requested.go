package handler

import (
	"judge-backend/internal/app/judge_domain/events/rejudgerejectionrequested"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeRejectionRequestedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterRejudgeRejectionRequestedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &RejudgeRejectionRequestedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[RejudgeRejectionRequestedHandler] Registered")
}

func (h *RejudgeRejectionRequestedHandler) Handle(e *sharedevent.RejudgeRejectionRequested) {
	h.logger.Debugf("[RejudgeRejectionRequestedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(rejudgerejectionrequested.New(e.ID(), e.RejudgeID()))
}
