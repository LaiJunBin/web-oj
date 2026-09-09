package handler

import (
	"errors"

	"judge-backend/internal/app/judge_domain/events/projectsubmitted"
	"judge-backend/internal/app/judge_domain/events/submissioncreated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	"judge-backend/internal/app/judge_domain/judge/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type ProjectSubmittedHandler struct {
	eventbus          sharedtypes.EventBus
	eventRepo         sharedtypes.EventRepository
	submissionRepo    repository.SubmissionRepository
	submissionService service.SubmissionService
	judgmentRepo      repository.JudgmentRepository
	judgmentFactory   *factory.JudgmentFactory
	logger            sharedtypes.Logger
}

func RegisterProjectSubmittedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	submissionRepo repository.SubmissionRepository,
	submissionService service.SubmissionService,
	judgmentRepo repository.JudgmentRepository,
	judgmentFactory *factory.JudgmentFactory,
	logger sharedtypes.Logger,
) {
	handler := &ProjectSubmittedHandler{
		eventbus:          eventbus,
		eventRepo:         eventRepo,
		submissionRepo:    submissionRepo,
		submissionService: submissionService,
		judgmentRepo:      judgmentRepo,
		judgmentFactory:   judgmentFactory,
		logger:            logger,
	}

	handler.eventbus.SubscribeAsync(handler.CreateSubmission, false)
	logger.Infof("[ProjectSubmittedHandler] Registered project submitted handler")
}

func (h *ProjectSubmittedHandler) CreateSubmission(e *projectsubmitted.Event) {
	h.logger.Debugf("[ProjectSubmittedHandler] Processing project submitted event for user %d, project %d", e.UserID(), e.ContestProjectID())

	claimed, err := h.eventRepo.TryClaimEvent(e.ID())
	if err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to claim event %s: %v", e.ID(), err)
		return
	}
	if !claimed {
		h.logger.Debugf("[ProjectSubmittedHandler] Event %s already processed, dropping", e.ID())
		return
	}

	submission, err := submission.New(&submission.SubmissionConfig{
		UserID:           e.UserID(),
		ContestProjectID: e.ContestProjectID(),
		GradingItemIds:   e.GradingItemIds(),
		IsPartial:        e.IsPartial(),
	})
	if err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to create submission: %v", err)
		return
	}

	// filesystem first: if these fail, DB is still clean and no compensation
	// needed. submission.New guarantees the ID exists before the writes below
	// store files under storage/submissions/{id}/.
	if err := h.submissionService.SaveProject(submission.ID(), e.Project()); err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to save project for submission %s: %v", submission.ID(), err)
		h.submissionService.CleanupSubmissionFiles(submission.ID())
		h.recordFailedSubmission(submission, "解壓縮或儲存提交檔案失敗")
		return
	}
	h.logger.Debugf("[ProjectSubmittedHandler] Saved project for submission %s", submission.ID())

	if err := h.submissionService.SaveDynamicFields(submission.ID(), e.DynamicFields()); err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to save dynamic fields for submission %s: %v", submission.ID(), err)
		h.submissionService.CleanupSubmissionFiles(submission.ID())
		h.recordFailedSubmission(submission, "儲存動態欄位失敗")
		return
	}
	h.logger.Debugf("[ProjectSubmittedHandler] Saved dynamic fields for submission %s", submission.ID())

	if err := h.submissionRepo.Save(submission); err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to save submission %s: %v", submission.ID(), err)
		h.submissionService.CleanupSubmissionFiles(submission.ID())
		// Best-effort visibility, same as the two failures above: yes, the
		// first step of recordFailedSubmission retries the very Save that
		// just failed — a transient error may clear, and if it does not, the
		// record path logs and gives up exactly like the SaveProject branch.
		h.recordFailedSubmission(submission, "儲存提交記錄失敗")
		return
	}
	h.logger.Debugf("[ProjectSubmittedHandler] Saved submission %s", submission.ID())

	h.eventbus.Publish(submissioncreated.NewSubmissionCreatedEvent(
		submission.ID(),
		submission.ContestProjectID(),
		submission.IsPartial()),
	)
	h.logger.Infof("[ProjectSubmittedHandler] Successfully published submission created event for submission %s", submission.ID())
}

// recordFailedSubmission persists the submission and attaches a terminal ERROR
// judgment, so a submission whose files could not be stored is VISIBLE to the
// contestant and staff as failed instead of vanishing — every list/detail query
// requires a judgment, so the row alone would stay hidden. This mirrors how
// OrphanSubmissionRepairService records the adjacent "saved but no judgment"
// failure; the uploaded project.zip is kept on disk (CleanupSubmissionFiles
// preserves it) for diagnosis. No retry is attempted here — durable redelivery
// across a process restart is a separate, deliberately-deferred concern.
func (h *ProjectSubmittedHandler) recordFailedSubmission(sub *submission.Submission, reason string) {
	if err := h.submissionRepo.Save(sub); err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to persist failed submission %s: %v", sub.ID(), err)
		return
	}

	judgeType := judgetype.JUDGE
	if sub.IsPartial() {
		judgeType = judgetype.JUDGE_PARTIAL
	}

	judgment, err := h.judgmentFactory.CreateJudgment(sub.ID(), judgeType)
	if err != nil {
		// The failure being recorded can coincide with an unresolvable
		// grading-item set; falling back to the item-less error judgment
		// keeps this submission visible instead of silently dropping it.
		if !errors.Is(err, repository.ErrGradingItemsNotOwned) {
			h.logger.Errorf("[ProjectSubmittedHandler] Failed to create error judgment for submission %s: %v", sub.ID(), err)
			return
		}
		judgment, err = h.judgmentFactory.CreateFailedJudgment(sub.ID(), judgeType, reason)
		if err != nil {
			h.logger.Errorf("[ProjectSubmittedHandler] Failed to create fallback error judgment for submission %s: %v", sub.ID(), err)
			return
		}
	} else {
		if err := judgment.Fail(); err != nil {
			h.logger.Errorf("[ProjectSubmittedHandler] Cannot mark submission %s judgment as failed: %v", sub.ID(), err)
			return
		}
		judgment.SetReason(reason)
		judgment.SetErrorMessage(reason)
	}

	if err := h.judgmentRepo.Save(judgment); err != nil {
		h.logger.Errorf("[ProjectSubmittedHandler] Failed to save error judgment for submission %s: %v", sub.ID(), err)
		return
	}

	h.logger.Infof("[ProjectSubmittedHandler] Recorded failed submission %s with ERROR judgment %s", sub.ID(), judgment.ID())
}
