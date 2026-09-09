package repository

import "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"

type ProjectRepository interface {
	FindBySubmissionID(submissionID string) (*project.Project, error)
}
