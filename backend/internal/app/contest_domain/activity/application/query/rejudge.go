package query

import "judge-backend/internal/app/contest_domain/activity/application/query/dto"

type RejudgeQuery interface {
	ListRejudges() ([]*dto.RejudgeInfo, error)
	ListRejudgesByContestProjectID(contestProjectID int32) ([]*dto.RejudgeInfo, error)
	GetRejudgeByID(rejudgeID string) (*dto.Rejudge, error)
	GetSubmissionComparison(rejudgeID string, submissionID string) (*dto.SubmissionComparisonDetail, error)
}
