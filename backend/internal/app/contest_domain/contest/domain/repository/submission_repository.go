package repository

import "judge-backend/internal/app/contest_domain/contest/domain/entity/submission"

type SubmissionRepository interface {
	FindByID(id string) (*submission.Submission, error)
}
