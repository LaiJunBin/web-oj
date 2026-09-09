package handler

import (
	"time"

	"judge-backend/internal/app/judge_domain/events/canceljudgerequested"
	"judge-backend/internal/app/judge_domain/events/judgecancelled"
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type CancelJudgeRequested struct {
	eventbus     sharedtypes.EventBus
	judgmentRepo repository.JudgmentRepository
	logger       sharedtypes.Logger
}

func RegisterCancelJudgeRequestedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	judgmentRepo repository.JudgmentRepository,
	logger sharedtypes.Logger,
) {
	handler := &CancelJudgeRequested{
		eventbus:     eventbus,
		judgmentRepo: judgmentRepo,
		logger:       logger,
	}

	// Claimed subscription: the handler is idempotent — TryCancel wins
	// at most once — so a transient failure releases the claim and retries
	// instead of silently discarding a cancellation the API already
	// confirmed.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.CancelJudge, sharedeventbus.ClaimedOptions[*canceljudgerequested.Event]{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[CancelJudgeRequested] Failed to subscribe: %v", err)
	}
	logger.Infof("[CancelJudgeRequested] Registered cancel judge requested handler")
}

func (h *CancelJudgeRequested) CancelJudge(e *canceljudgerequested.Event) error {
	h.logger.Debugf("[CancelJudgeRequested] Processing cancel judge requested event for judgment %s", e.JudgmentID())

	// TryCancel, not FindByID + mutate + Save: a FindByID'd snapshot can go
	// stale the instant the executing worker writes progress this snapshot
	// never saw (grading results, testcase statuses) — saving the whole
	// document back would revert that progress. TryCancel only ever touches
	// status/reason/judgingFinishedAt, and is itself the guard against
	// racing the worker's own terminal save (see its doc comment). It also
	// returns the post-update judgment from the same atomic command, so the
	// events below never depend on a separate re-read that could fail after
	// the cancellation was already persisted.
	cancelled, err := h.judgmentRepo.TryCancel(e.JudgmentID(), e.Reason(), time.Now())
	if err != nil {
		// Returning the error hands the event back to the claimed
		// subscription for a bounded retry — this exact failure used to be
		// the old silent-drop hole: the API had already said OK and nobody would ever
		// re-issue the cancel.
		h.logger.Errorf("[CancelJudgeRequested] Failed to cancel judgment %s: %v", e.JudgmentID(), err)
		return err
	}
	if cancelled == nil {
		// Benign no-op, not a failure: retrying cannot change a terminal
		// status.
		h.logger.Debugf("[CancelJudgeRequested] Judgment %s already reached a terminal status, skipping cancellation", e.JudgmentID())
		return nil
	}
	h.logger.Debugf("[CancelJudgeRequested] Set judgment %s status to CANCELLED with reason: %s", e.JudgmentID(), e.Reason())

	h.eventbus.Publish(judgecancelled.NewJudgeCancelledEvent(e.JudgmentID()))
	h.logger.Debugf("[CancelJudgeRequested] Published judge cancelled event for judgment %s", e.JudgmentID())

	// As the writer that won the Cancelled transition, this handler must also
	// publish the judgeupdated the rest of the system keys off. When the
	// judgment was still QUEUED, nobody else ever will: no worker is
	// executing it, and its later TryStart just fails silently — without
	// this, the rejudge rollup never sees the member reach a terminal
	// status (the rejudge hangs in RUNNING forever) and the contest stream
	// never shows the cancellation. When a worker IS executing it, that
	// worker publishes its own judgeupdated after losing the terminal save;
	// the duplicate is harmless (the rollup's transitions are conditional
	// and the stream update is idempotent).
	h.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(cancelled))
	h.logger.Debugf("[CancelJudgeRequested] Published judge updated event for cancelled judgment %s", e.JudgmentID())
	return nil
}
