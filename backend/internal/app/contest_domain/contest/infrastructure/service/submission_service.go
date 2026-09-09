package service

import (
	"judge-backend/internal/app/contest_domain/contest/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/contest/domain/service"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type SubmissionService struct {
	eventBus sharedtypes.EventBus
	logger   sharedtypes.Logger
}

func NewSubmissionService(
	eventBus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) service.SubmissionService {
	return &SubmissionService{
		eventBus: eventBus,
		logger:   logger,
	}
}

func (s *SubmissionService) RequestCancelSubmission(submission *submission.Submission) {
	s.logger.Debugf("Starting to request cancel submission", "submissionId", submission.ID(), "judgmentId", submission.JudgmentID())
	s.eventBus.Publish(sharedevent.NewCancelJudgeRequested(submission.JudgmentID(), "cancel by user"))
	s.logger.Debugf("Successfully published cancel submission event", "submissionId", submission.ID(), "judgmentId", submission.JudgmentID())
}
