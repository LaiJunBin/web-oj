package service

import "judge-backend/internal/app/contest_domain/contest/domain/entity/submission"

type SubmissionService interface {
	RequestCancelSubmission(submission *submission.Submission)
}
