package query

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"

	"gorm.io/gorm"
)

type LifecycleQuery struct {
	db *gorm.DB
}

func NewLifecycleQuery(db *gorm.DB) query.LifecycleQuery {
	return &LifecycleQuery{db: db}
}

func (q *LifecycleQuery) GetLifecycle(id int32) (*dto.LifecycleCommand, error) {
	lifecycle := &model.Lifecycle{}
	if err := q.db.Where("id = ?", id).First(&lifecycle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &dto.LifecycleCommand{
		ID:                   lifecycle.ID,
		Type:                 lifecycle.Type,
		ExecutionEnvironment: lifecycle.RunsOn,
		Name:                 lifecycle.Name,
		Description:          lifecycle.Description,
		Command:              lifecycle.Command,
		CreatedAt:            lifecycle.CreatedAt,
		UpdatedAt:            lifecycle.UpdatedAt,
	}, nil
}
