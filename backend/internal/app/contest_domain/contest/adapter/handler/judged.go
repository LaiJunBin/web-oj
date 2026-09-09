package handler

import (
	"judge-backend/internal/app/contest_domain/events/submissionjudged"
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

func (h *JudgedHandler) Handle(e *sharedevent.Judged) {
	h.logger.Debugf("[JudgedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(submissionjudged.New(e.ID(), e.ContestProjectID()))
}
