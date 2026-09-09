package repository

import "judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"

type LifecycleRepository interface {
	FindByID(id int) (*lifecycle.Lifecycle, error)
	Save(lifecycle *lifecycle.Lifecycle) error
	Delete(lifecycle *lifecycle.Lifecycle) error
}
