package handler

import (
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/events/staffnoteupdaterequested"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type StaffNoteUpdateRequested struct {
	eventbus     sharedtypes.EventBus
	judgmentRepo repository.JudgmentRepository
	logger       sharedtypes.Logger
}

func RegisterStaffNoteUpdateRequestedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	judgmentRepo repository.JudgmentRepository,
	logger sharedtypes.Logger,
) {
	handler := &StaffNoteUpdateRequested{
		eventbus:     eventbus,
		judgmentRepo: judgmentRepo,
		logger:       logger,
	}

	// Claimed subscription, like cancellation: the write is a plain $set of
	// one field, so a retry after a transient failure is harmless, while a
	// silent drop would leave the admin believing a note the contestant
	// never sees was saved.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.UpdateStaffNote, sharedeventbus.ClaimedOptions[*staffnoteupdaterequested.Event]{}); err != nil {
		logger.Panicf("[StaffNoteUpdateRequested] Failed to subscribe: %v", err)
	}
	logger.Infof("[StaffNoteUpdateRequested] Registered staff note update requested handler")
}

func (h *StaffNoteUpdateRequested) UpdateStaffNote(e *staffnoteupdaterequested.Event) error {
	h.logger.Debugf("[StaffNoteUpdateRequested] Processing staff note update for judgment %s", e.JudgmentID())

	// UpdateStaffNote, never FindByID + mutate + Save: the note may be
	// written while a worker is still executing the judgment, and saving a
	// whole snapshot back would revert progress the snapshot never saw.
	// Last write wins by design — there is no author or history to merge.
	updated, err := h.judgmentRepo.UpdateStaffNote(e.JudgmentID(), e.StaffNote())
	if err != nil {
		h.logger.Errorf("[StaffNoteUpdateRequested] Failed to update staff note of judgment %s: %v", e.JudgmentID(), err)
		return err
	}
	if updated == nil {
		// The judgment vanished between the API's existence check and this
		// write; retrying cannot bring it back.
		h.logger.Warnf("[StaffNoteUpdateRequested] Judgment %s not found, skipping staff note update", e.JudgmentID())
		return nil
	}

	// The same judgeupdated the judge pipeline publishes, so the contestant's
	// open stream (and any other listener) picks the note up without a
	// dedicated event type.
	h.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(updated))
	h.logger.Debugf("[StaffNoteUpdateRequested] Published judge updated event for judgment %s", e.JudgmentID())
	return nil
}
