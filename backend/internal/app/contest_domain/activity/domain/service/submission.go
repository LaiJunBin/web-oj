package service

import "judge-backend/internal/app/contest_domain/activity/domain/entity/submission"

type SubmissionService interface {
	RequestRejudgeSubmission(submission *submission.Submission, reason string)

	// RequestRejudgeSubmissions takes the contest project separately rather than
	// reading it off the submissions: a project with no submissions at all is
	// still a legitimate rejudge request, and the resulting rejudge has to record
	// which project it was raised against.
	RequestRejudgeSubmissions(contestProjectID int32, submissions []*submission.Submission, reason string)
}
