package handler

import (
	"judge-backend/internal/app/judge_domain/events/cancelrejudgerequested"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type CancelRejudgeRequestedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterCancelRejudgeRequestedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &CancelRejudgeRequestedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[CancelRejudgeRequestedHandler] Registered")
}

func (h *CancelRejudgeRequestedHandler) Handle(e *sharedevent.CancelRejudgeRequested) {
	h.logger.Debugf("[CancelRejudgeRequestedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(cancelrejudgerequested.New(e.ID(), e.RejudgeID()))
}
