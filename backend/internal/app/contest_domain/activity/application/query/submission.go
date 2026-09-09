package query

import "judge-backend/internal/app/contest_domain/activity/application/query/dto"

type SubmissionQuery interface {
	GetSubmissionsByContestProjectID(contestProjectID int32) ([]*dto.Submission, error)
	GetSubmissionInfoByID(submissionID string) (*dto.Submission, error)
	GetSubmissionByID(submissionID string) (*dto.SubmissionDetail, error)
	GetSubmissionsByUserID(userID int32) ([]*dto.Submission, error)
}
