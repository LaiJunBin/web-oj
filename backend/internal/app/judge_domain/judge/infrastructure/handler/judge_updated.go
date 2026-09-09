package handler

import (
	"fmt"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgeUpdatedHandler struct {
	rejudgeRepo repository.RejudgeRepository
	logger      sharedtypes.Logger
}

func RegisterJudgeUpdatedHandler(
	eventbus sharedtypes.EventBus,
	rejudgeRepo repository.RejudgeRepository,
	logger sharedtypes.Logger,
) {
	handler := &JudgeUpdatedHandler{
		rejudgeRepo: rejudgeRepo,
		logger:      logger,
	}

	// Retried, NOT claimed: this handler must see EVERY judgeupdated
	// delivery — a claim would swallow the later events of the same judgment,
	// and the terminal one carries the FINISHED/ERROR flip. What it needs is
	// only that one transient Mongo error does not permanently strand the
	// rejudge in RUNNING (that flip is published exactly once); every
	// transition below is conditional, so a retry re-running from the top
	// converges.
	if err := sharedeventbus.SubscribeRetriedAsync(eventbus, logger, false,
		func(e *judgeupdated.JudgeUpdatedEvent) string { return e.ID() },
		handler.UpdateRejudgeStatus, sharedeventbus.RetryOptions{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[JudgeUpdatedHandler] Failed to subscribe: %v", err)
	}
	logger.Infof("[JudgeUpdatedHandler] Registered judge updated handler")
}

// UpdateRejudgeStatus deliberately does NOT claim/dedupe the event: a
// judgment publishes many judgeupdated events over its life (their IDs are
// unique per publish), the in-process bus delivers each exactly once, and
// this handler must see every one of them — the terminal event carries the
// FINISHED/ERROR flip.
//
// Every status change below goes through TransitionStatus (a conditional DB
// update), never a read-then-Save cycle: this handler runs on its own
// goroutine per event (SubscribeAsync(..., false) — no serialization), and
// multiple judgments in the same rejudge batch (or multiple events from the
// SAME judgment) can invoke it concurrently. A stale in-memory read followed
// by an unconditional Save would let an earlier, slower update clobber a
// later, terminal one back to an earlier status.
//
// A returned error hands the event back to the retried subscription for a
// bounded re-run; the retry core owns the logging. applied=false stays
// discarded — a lost conditional race is normal, not a failure.
func (h *JudgeUpdatedHandler) UpdateRejudgeStatus(event *judgeupdated.JudgeUpdatedEvent) error {
	d := event.Data()
	h.logger.Debugf("[JudgeUpdatedHandler] Processing judge updated event for judgment %s", d.JudgmentID)

	if d.RejudgeID == nil {
		return nil
	}
	rejudgeID := *d.RejudgeID
	notYetTerminal := []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}

	if _, err := h.rejudgeRepo.TransitionStatus(rejudgeID, []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING); err != nil {
		return fmt.Errorf("transition rejudge %s to RUNNING: %w", rejudgeID, err)
	}

	if d.StatusIsError {
		if _, err := h.rejudgeRepo.TransitionStatus(rejudgeID, notYetTerminal, rejudge.ERROR); err != nil {
			return fmt.Errorf("transition rejudge %s to ERROR: %w", rejudgeID, err)
		}
	}

	if d.StatusIsCancelled {
		if _, err := h.rejudgeRepo.TransitionStatus(rejudgeID, notYetTerminal, rejudge.CANCELED); err != nil {
			return fmt.Errorf("transition rejudge %s to CANCELED: %w", rejudgeID, err)
		}
	}

	if d.StatusIsRejudged {
		allFinished, err := h.rejudgeRepo.AllJudgmentsFinished(rejudgeID)
		if err != nil {
			return fmt.Errorf("check all judgments finished for rejudge %s: %w", rejudgeID, err)
		}
		if allFinished {
			if _, err := h.rejudgeRepo.TransitionStatus(rejudgeID, notYetTerminal, rejudge.FINISHED); err != nil {
				return fmt.Errorf("transition rejudge %s to FINISHED: %w", rejudgeID, err)
			}
		}
	}
	return nil
}
