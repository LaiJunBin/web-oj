package repository

import "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"

type LifecycleRepository interface {
	FindBySubmissionID(submissionID string) ([]*lifecycle.Lifecycle, error)
}
