package repository

import "judge-backend/internal/app/contest_domain/activity/domain/entity/submission"

type SubmissionRepository interface {
	FindByID(id string) (*submission.Submission, error)
	FindByContestProjectIDAndGradingItemIDs(contestProjectID int32, gradingItemIDs []int32) ([]*submission.Submission, error)
	Save(submission *submission.Submission) error
}
