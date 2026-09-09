package query

import "judge-backend/internal/app/contest_domain/contest/application/query/dto"

type SubmissionQuery interface {
	ListUserProjectSubmissionsByProjectID(userID, contestProjectID int32) ([]*dto.Submission, error)
	GetUserSubmissionByID(userID int32, submissionID string) (*dto.SubmissionDetail, error)
	// GetSampleTestcaseIDs re-resolves which of the contest project's
	// testcases are samples right now. The live submission stream calls it
	// (behind a short-TTL memo) so an already-open stream follows mid-contest
	// visibility changes instead of freezing the set it saw at open.
	GetSampleTestcaseIDs(contestProjectID int32) ([]int32, error)
}
