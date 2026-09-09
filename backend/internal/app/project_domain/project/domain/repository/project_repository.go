package repository

import "judge-backend/internal/app/project_domain/project/domain/entity/project"

type ProjectRepository interface {
	Exists(id int) (bool, error)
	FindByID(id int) (*project.Project, error)
	Save(project *project.Project) error
	Delete(project *project.Project) error
}
