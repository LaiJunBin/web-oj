package repository

import "judge-backend/internal/app/contest_domain/contest/domain/entity/project"

type ProjectRepository interface {
	FindByID(contestProjectID int32) (*project.Project, error)
}
