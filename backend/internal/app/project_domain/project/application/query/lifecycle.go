package query

import "judge-backend/internal/app/project_domain/project/application/query/dto"

type LifecycleQuery interface {
	GetLifecycle(id int32) (*dto.LifecycleCommand, error)
}
