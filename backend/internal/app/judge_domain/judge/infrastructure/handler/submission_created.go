package handler

import (
	"errors"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/events/judgmentcreated"
	"judge-backend/internal/app/judge_domain/events/submissioncreated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type SubmissionCreatedHandler struct {
	eventbus        sharedtypes.EventBus
	eventRepo       sharedtypes.EventRepository
	judgmentFactory *factory.JudgmentFactory
	judgmentRepo    repository.JudgmentRepository
	logger          sharedtypes.Logger
}

func RegisterSubmissionCreatedHandler(
	eventbus sharedtypes.EventBus,
	eventRepo sharedtypes.EventRepository,
	judgmentFactory *factory.JudgmentFactory,
	judgmentRepo repository.JudgmentRepository,
	logger sharedtypes.Logger,
) {
	handler := &SubmissionCreatedHandler{
		eventbus:        eventbus,
		eventRepo:       eventRepo,
		judgmentFactory: judgmentFactory,
		judgmentRepo:    judgmentRepo,
		logger:          logger,
	}

	handler.eventbus.SubscribeAsync(handler.CreateJudgment, false)
	logger.Infof("[SubmissionCreatedHandler] Registered submission created handler")
}

func (h *SubmissionCreatedHandler) CreateJudgment(event *submissioncreated.SubmissionCreatedEvent) {
	h.logger.Debugf("[SubmissionCreatedHandler] Processing submission created event for submission %s", event.SubmissionID())

	claimed, err := h.eventRepo.TryClaimEvent(event.ID())
	if err != nil {
		h.logger.Errorf("[SubmissionCreatedHandler] Failed to claim event %s: %v", event.ID(), err)
		return
	}
	if !claimed {
		h.logger.Debugf("[SubmissionCreatedHandler] Event %s already processed, dropping", event.ID())
		return
	}

	judgeType := judgetype.JUDGE
	if event.IsPartial() {
		judgeType = judgetype.JUDGE_PARTIAL
		h.logger.Debugf("[SubmissionCreatedHandler] Using partial judge type for submission %s", event.SubmissionID())
	}

	judgment, err := h.judgmentFactory.CreateJudgment(event.SubmissionID(), judgeType)
	if err != nil {
		if errors.Is(err, repository.ErrGradingItemsNotOwned) {
			h.recordUnresolvableSubmission(event.SubmissionID(), judgeType)
			return
		}
		h.logger.Errorf("[SubmissionCreatedHandler] Failed to create judgment for submission %s: %v", event.SubmissionID(), err)
		return
	}
	h.logger.Debugf("[SubmissionCreatedHandler] Created judgment %s for submission %s", judgment.ID(), event.SubmissionID())

	if err := h.judgmentRepo.Save(judgment); err != nil {
		h.logger.Errorf("[SubmissionCreatedHandler] Failed to save judgment %s: %v", judgment.ID(), err)
		return
	}
	h.logger.Debugf("[SubmissionCreatedHandler] Saved judgment %s", judgment.ID())

	h.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(judgment))
	h.logger.Debugf("[SubmissionCreatedHandler] Published judge updated event for judgment %s", judgment.ID())

	h.eventbus.Publish(judgmentcreated.NewJudgmentCreatedEvent(judgment))
	h.logger.Infof("[SubmissionCreatedHandler] Successfully published judgment created event for judgment %s", judgment.ID())
}

// recordUnresolvableSubmission persists a terminal ERROR judgment for a
// submission whose grading-item set does not resolve to its own problem.
// Without it the submission simply never got a judgment — invisible to the
// contestant and retried by the orphan repair sweep forever, since the
// failure is deterministic. JudgeUpdated makes it visible immediately;
// JudgmentCreated is deliberately NOT published, because that would enqueue
// the judgment and an ERROR judgment must never reach a worker.
func (h *SubmissionCreatedHandler) recordUnresolvableSubmission(submissionID string, judgeType judgetype.JudgeType) {
	failed, err := h.judgmentFactory.CreateFailedJudgment(submissionID, judgeType, factory.GradingItemMismatchReason)
	if err != nil {
		h.logger.Errorf("[SubmissionCreatedHandler] Failed to create error judgment for submission %s: %v", submissionID, err)
		return
	}
	if err := h.judgmentRepo.Save(failed); err != nil {
		h.logger.Errorf("[SubmissionCreatedHandler] Failed to save error judgment %s: %v", failed.ID(), err)
		return
	}
	h.eventbus.Publish(judgeupdated.NewJudgeUpdatedEvent(failed))
	h.logger.Infof("[SubmissionCreatedHandler] Recorded submission %s with ERROR judgment %s: unresolvable grading items", submissionID, failed.ID())
}
