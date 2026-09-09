package handler

import (
	"judge-backend/internal/app/judge_domain/events/staffnoteupdaterequested"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type StaffNoteUpdateRequestedHandler struct {
	eventbus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func RegisterStaffNoteUpdateRequestedHandler(
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	h := &StaffNoteUpdateRequestedHandler{eventbus: eventbus, logger: logger}
	eventbus.SubscribeAsync(h.Handle, false)
	logger.Infof("[StaffNoteUpdateRequestedHandler] Registered")
}

func (h *StaffNoteUpdateRequestedHandler) Handle(e *sharedevent.StaffNoteUpdateRequested) {
	h.logger.Debugf("[StaffNoteUpdateRequestedHandler] Translating shared event %s", e.ID())
	h.eventbus.Publish(staffnoteupdaterequested.New(e.ID(), e.JudgmentID(), e.StaffNote()))
}
