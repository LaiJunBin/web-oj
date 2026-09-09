package handler

import (
	"errors"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/events/judgmentcreated"
	"judge-backend/internal/app/judge_domain/events/rejudgerequested"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeRequested struct {
	eventbus        sharedtypes.EventBus
	eventRepo       sharedtypes.EventRepository
	judgmentFactory *factory.JudgmentFactory
	judgmentRepo    repository.JudgmentRepository
	rejudgeRepo     repository.RejudgeRepository
	logger          sharedtypes.Logger
}

func RegisterRejudgeRequestedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	judgmentFactory *factory.JudgmentFactory,
	judgmentRepo repository.JudgmentRepository,
	rejudgeRepo repository.RejudgeRepository,
	logger sharedtypes.Logger,
) {
	handler := &RejudgeRequested{
		eventbus:        eventbus,
		eventRepo:       eventRepo,
		judgmentFactory: judgmentFactory,
		judgmentRepo:    judgmentRepo,
		rejudgeRepo:     rejudgeRepo,
		logger:          logger,
	}

	handler.eventbus.SubscribeAsync(handler.CreateRejudgeJudgments, false)
	logger.Infof("[RejudgeRequested] Registered rejudge requested handler")
}

func (h *RejudgeRequested) CreateRejudgeJudgments(e *rejudgerequested.Event) {
	h.logger.Debugf("[RejudgeRequested] Processing rejudge requested event %s for %d submissions", e.ID(), len(e.SubmissionIDs()))

	claimed, err := h.eventRepo.TryClaimEvent(e.ID())
	if err != nil {
		h.logger.Errorf("[RejudgeRequested] Failed to claim event %s: %v", e.ID(), err)
		return
	}
	if !claimed {
		h.logger.Debugf("[RejudgeRequested] Event %s already processed, dropping", e.ID())
		return
	}

	rejudge := rejudge.New(&rejudge.RejudgeConfig{
		ID:               e.ID(),
		ContestProjectID: e.ContestProjectID(),
		Reason:           e.Reason(),
		Status:           rejudge.PENDING,
		Judgments:        make([]string, 0),
	})
	h.logger.Infof("[RejudgeRequested] Created new rejudge %s with reason: %s", rejudge.ID(), rejudge.Reason())

	// Create and persist all judgments first, WITHOUT queueing them: a queued
	// judgment can finish at any moment after publish, and its JudgeUpdated
	// handler must find a fully persisted rejudge — otherwise the status flip
	// is lost or overwritten by the save below.
	judgments := make([]*judgment.Judgment, 0, len(e.SubmissionIDs()))
	failedJudgments := make([]*judgment.Judgment, 0)
	for _, submissionID := range e.SubmissionIDs() {
		h.logger.Debugf("[RejudgeRequested] Creating judgment for submission %s", submissionID)
		j, err := h.judgmentFactory.CreateJudgment(submissionID, judgetype.REJUDGE)
		if err != nil {
			// An unresolvable grading-item set still gets a judgment — a
			// terminal ERROR one, attached to this rejudge but never queued.
			// Skipping instead used to leave the rejudge with fewer (possibly
			// zero) judgments, and with zero it sat in PENDING forever: its
			// status only ever advances on judge-updated events.
			if errors.Is(err, repository.ErrGradingItemsNotOwned) {
				if failed := h.recordUnresolvableRejudgeJudgment(submissionID, rejudge); failed != nil {
					failedJudgments = append(failedJudgments, failed)
				}
				continue
			}
			h.logger.Errorf("[RejudgeRequested] Failed to create judgment for submission %s: %v", submissionID, err)
			continue
		}

		j.SetRejudgeID(rejudge.ID())
		h.logger.Debugf("[RejudgeRequested] Set rejudge ID %s for judgment %s", rejudge.ID(), j.ID())

		if err := h.judgmentRepo.Save(j); err != nil {
			h.logger.Errorf("[RejudgeRequested] Failed to save judgment %s: %v", j.ID(), err)
			continue
		}
		h.logger.Debugf("[RejudgeRequested] Saved judgment %s", j.ID())

		rejudge.AddJudgmentID(j.ID())
		judgments = append(judgments, j)
	}

	if err := h.rejudgeRepo.Save(rejudge); err != nil {
		h.logger.Errorf("[RejudgeRequested] Failed to save rejudge %s: %v", rejudge.ID(), err)
		return
	}
	h.logger.Infof("[RejudgeRequested] Successfully saved rejudge %s with %d judgments", rejudge.ID(), len(rejudge.Judgments()))

	// Only now hand the judgments to the queue.
	for _, j := range judgments {
		h.eventbus.Publish(judgmentcreated.NewJudgmentCreatedEvent(j))
		h.logger.Debugf("[RejudgeRequested] Published judgment created event for judgment %s", j.ID())
	}

	// A judgment that failed before it could run still has to reach the
	// rejudge status machinery: publishing its JudgeUpdated AFTER the rejudge
	// is persisted (the same ordering rule as above) walks the rejudge through
	// the existing error path, exactly as an execution-time failure would.
	// These are never published as JudgmentCreated — nothing to queue.
	for _, j := range failedJudgments {
		h.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(j))
		h.logger.Debugf("[RejudgeRequested] Published judge updated event for failed judgment %s", j.ID())
	}
}

// recordUnresolvableRejudgeJudgment persists the terminal ERROR judgment for
// one submission of a rejudge whose grading items no longer resolve, and
// registers it on the rejudge. Returns nil when even that could not be saved.
func (h *RejudgeRequested) recordUnresolvableRejudgeJudgment(submissionID string, r *rejudge.Rejudge) *judgment.Judgment {
	failed, err := h.judgmentFactory.CreateFailedJudgment(submissionID, judgetype.REJUDGE, factory.GradingItemMismatchReason)
	if err != nil {
		h.logger.Errorf("[RejudgeRequested] Failed to create error judgment for submission %s: %v", submissionID, err)
		return nil
	}
	failed.SetRejudgeID(r.ID())
	if err := h.judgmentRepo.Save(failed); err != nil {
		h.logger.Errorf("[RejudgeRequested] Failed to save error judgment %s: %v", failed.ID(), err)
		return nil
	}
	r.AddJudgmentID(failed.ID())
	h.logger.Infof("[RejudgeRequested] Recorded submission %s with ERROR judgment %s: unresolvable grading items", submissionID, failed.ID())
	return failed
}
