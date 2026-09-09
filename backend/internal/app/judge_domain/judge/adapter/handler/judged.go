package handler

import (
	"judge-backend/internal/app/judge_domain/events/judged"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterJudgedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &JudgedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[JudgedHandler] Registered")
}

func (h *JudgedHandler) Handle(e *judged.JudgedEvent) {
	h.logger.Debugf("[JudgedHandler] Translating judged event %s", e.ID())
	h.eventbus.Publish(sharedevent.NewJudged(e.JudgmentID(), e.ContestProjectID()))
}
