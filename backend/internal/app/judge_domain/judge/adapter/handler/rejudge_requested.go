package handler

import (
	"judge-backend/internal/app/judge_domain/events/rejudgerequested"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeRequestedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterRejudgeRequestedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &RejudgeRequestedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[RejudgeRequestedHandler] Registered")
}

func (h *RejudgeRequestedHandler) Handle(e *sharedevent.RejudgeRequested) {
	h.logger.Debugf("[RejudgeRequestedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(rejudgerequested.New(e.ID(), e.ContestProjectID(), e.SubmissionIDs(), e.Reason()))
}
