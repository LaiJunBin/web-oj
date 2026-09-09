package handler

import (
	"judge-backend/internal/app/judge_domain/events/projectsubmitted"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type ProjectSubmittedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterProjectSubmittedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &ProjectSubmittedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[ProjectSubmittedHandler] Registered")
}

func (h *ProjectSubmittedHandler) Handle(e *sharedevent.ProjectSubmitted) {
	h.logger.Debugf("[ProjectSubmittedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(projectsubmitted.New(
		e.ID(),
		e.UserID(),
		e.ContestProjectID(),
		e.IsPartial(),
		e.GradingItemIds(),
		e.Project(),
		e.DynamicFields(),
		e.CreatedAt(),
	))
}
