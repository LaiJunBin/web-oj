package service

import (
	"judge-backend/internal/app/contest_domain/activity/domain/entity/submission"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type SubmissionService struct {
	eventBus sharedtypes.EventBus
}

func NewSubmissionService(eventBus sharedtypes.EventBus) service.SubmissionService {
	return &SubmissionService{
		eventBus: eventBus,
	}
}

func (s *SubmissionService) RequestRejudgeSubmission(submission *submission.Submission, reason string) {
	s.eventBus.Publish(sharedevent.NewRejudgeRequested(submission.ContestProjectID(), []string{submission.ID()}, reason))
}

func (s *SubmissionService) RequestRejudgeSubmissions(contestProjectID int32, submissions []*submission.Submission, reason string) {
	submissionIDs := make([]string, 0, len(submissions))
	for _, sub := range submissions {
		submissionIDs = append(submissionIDs, sub.ID())
	}
	s.eventBus.Publish(sharedevent.NewRejudgeRequested(contestProjectID, submissionIDs, reason))
}
