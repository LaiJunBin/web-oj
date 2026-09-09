package handler

import (
	"fmt"
	"time"

	"judge-backend/internal/app/judge_domain/events/cancelrejudgerequested"
	"judge-backend/internal/app/judge_domain/events/judgecancelled"
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type CancelRejudgeRequested struct {
	eventbus     sharedtypes.EventBus
	rejudgeRepo  repository.RejudgeRepository
	judgmentRepo repository.JudgmentRepository
	logger       sharedtypes.Logger
}

func RegisterCancelRejudgeRequestedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	rejudgeRepo repository.RejudgeRepository,
	judgmentRepo repository.JudgmentRepository,
	logger sharedtypes.Logger,
) {
	handler := &CancelRejudgeRequested{
		eventbus:     eventbus,
		rejudgeRepo:  rejudgeRepo,
		judgmentRepo: judgmentRepo,
		logger:       logger,
	}

	// Claimed subscription: every step is conditional (TransitionStatus,
	// per-member TryCancel), so the whole handler converges under retry —
	// this automates the "press cancel again to heal stragglers" repair that
	// reconcileMembers' doc comment used to ask of the user.
	if err := sharedeventbus.SubscribeClaimedAsync(eventbus, eventRepo, logger, false,
		handler.CancelRejudge, sharedeventbus.ClaimedOptions[*cancelrejudgerequested.Event]{}); err != nil {
		// Startup-time programming error (the handler is not a valid func):
		// the app must not come up half-wired.
		logger.Panicf("[CancelRejudgeRequested] Failed to subscribe: %v", err)
	}
	logger.Infof("[CancelRejudgeRequested] Registered cancel rejudge requested handler")
}

func (h *CancelRejudgeRequested) CancelRejudge(e *cancelrejudgerequested.Event) error {
	h.logger.Debugf("[CancelRejudgeRequested] Processing cancel rejudge requested event for rejudge %s", e.RejudgeID())

	// Judgments() is populated once at rejudge creation and never mutated
	// afterward, so it's safe to read regardless of when the status
	// transition below lands. Read it BEFORE transitioning, not after: if we
	// flipped status to CANCELED first and this lookup failed afterward, the
	// rejudge would be left showing CANCELED in the DB/UI while none of its
	// judgments ever received a cancel event — workers would keep running
	// them to completion with no compensation until a retry lands. Failing
	// here (returning the error for a bounded retry) leaves the rejudge in
	// whatever state it already was, which is the safe direction to fail in.
	r, err := h.rejudgeRepo.FindByID(e.RejudgeID())
	if err != nil {
		h.logger.Errorf("[CancelRejudgeRequested] Failed to find rejudge %s: %v", e.RejudgeID(), err)
		return err
	}

	// Conditional update, not a read-then-Save: this can race the automatic
	// per-judgment status rollup in JudgeUpdatedHandler on the same rejudge
	// document. See TransitionStatus's doc comment.
	applied, err := h.rejudgeRepo.TransitionStatus(e.RejudgeID(), []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED)
	if err != nil {
		h.logger.Errorf("[CancelRejudgeRequested] Failed to cancel rejudge %s: %v", e.RejudgeID(), err)
		return err
	}
	if !applied {
		// Not necessarily "nothing to do": a previous cancellation attempt
		// may have flipped the parent to CANCELED and then died partway
		// through its members, leaving Queued/Judging children under a
		// CANCELED parent. Member reconciliation below is idempotent
		// (TryCancel wins at most once per member), so when the parent's
		// CURRENT status is CANCELED we reconcile again instead of
		// returning — this is what makes a re-received cancel request (the
		// user pressing cancel again, or any future redelivery) heal
		// stragglers rather than silently doing nothing. Re-read the
		// status: r's copy predates the transition attempt and may be stale.
		current, err := h.rejudgeRepo.FindByID(e.RejudgeID())
		if err != nil {
			h.logger.Errorf("[CancelRejudgeRequested] Failed to re-read rejudge %s after a not-applied transition: %v", e.RejudgeID(), err)
			return err
		}
		if !current.Status().IsCanceled() {
			// Benign: the rejudge finished/was decided first; nothing a
			// retry could change.
			h.logger.Debugf("[CancelRejudgeRequested] Rejudge %s was no longer cancellable (status %s), skipping", e.RejudgeID(), current.Status().String())
			return nil
		}
		h.logger.Infof("[CancelRejudgeRequested] Rejudge %s is already CANCELED; reconciling its members", e.RejudgeID())
	} else {
		h.logger.Infof("[CancelRejudgeRequested] Successfully set rejudge %s status to CANCELED", e.RejudgeID())
	}

	return h.reconcileMembers(r)
}

// reconcileMembers cancels every member of an already-CANCELED rejudge in
// the DATABASE first, then publishes events only for the members this call
// actually transitioned. Publishing judge_cancelled alone (the old behavior)
// only interrupted members a worker happened to be executing right now —
// every still-QUEUED member kept its Queued status, was later popped and
// executed to completion, with the pre-published events' claims (keyed by
// judgment ID) swallowing any legitimate cancel of the same judgment for the
// claim TTL.
//
// The loop is idempotent: members already terminal report (nil, nil) and get
// no events, so running it any number of times converges. A member whose
// TryCancel errors is counted and reported as an error at the end, which
// hands the whole event back for a bounded retry — the in-process
// automation of the "re-request the cancel to heal stragglers" repair
// (in-process only: if the process dies, re-requesting still works).
// TryCancel with an empty reason keeps whatever reason each judgment already
// has, matching what the worker's own ctx-cancelled path records.
func (h *CancelRejudgeRequested) reconcileMembers(r *rejudge.Rejudge) error {
	failed := 0
	for _, judgmentID := range r.Judgments() {
		cancelled, err := h.judgmentRepo.TryCancel(judgmentID, "", time.Now())
		if err != nil {
			h.logger.Errorf("[CancelRejudgeRequested] Failed to cancel judgment %s of rejudge %s: %v", judgmentID, r.ID(), err)
			failed++
			continue
		}
		if cancelled == nil {
			h.logger.Debugf("[CancelRejudgeRequested] Judgment %s of rejudge %s already reached a terminal status, skipping", judgmentID, r.ID())
			continue
		}

		// Interrupts the worker currently executing this member, if any.
		h.eventbus.Publish(judgecancelled.NewJudgeCancelledEvent(judgmentID))

		// As the writer that won this member's Cancelled transition, this
		// handler also publishes its judgeupdated — for queued members
		// nobody else ever will (see CancelJudgeRequested for the full
		// rationale). The entity comes from TryCancel's own atomic command,
		// so a won transition always has its event payload.
		h.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(cancelled))
	}
	h.logger.Debugf("[CancelRejudgeRequested] Finished reconciling judgments of rejudge %s", r.ID())
	if failed > 0 {
		return fmt.Errorf("%d of %d members of rejudge %s failed to cancel", failed, len(r.Judgments()), r.ID())
	}
	return nil
}
