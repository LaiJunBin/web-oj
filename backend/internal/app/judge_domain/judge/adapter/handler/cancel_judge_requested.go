package handler

import (
	"judge-backend/internal/app/judge_domain/events/canceljudgerequested"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type CancelJudgeRequestedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterCancelJudgeRequestedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &CancelJudgeRequestedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[CancelJudgeRequestedHandler] Registered")
}

func (h *CancelJudgeRequestedHandler) Handle(e *sharedevent.CancelJudgeRequested) {
	h.logger.Debugf("[CancelJudgeRequestedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(canceljudgerequested.New(e.ID(), e.JudgmentID(), e.Reason()))
}
