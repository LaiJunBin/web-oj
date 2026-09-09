package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/domain/entity/lifecycle"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"

	"gorm.io/gorm"
)

type LifecycleRepository struct {
	db *gorm.DB
}

func NewLifecycleRepository(db *gorm.DB) repository.LifecycleRepository {
	return &LifecycleRepository{
		db: db,
	}
}

func (r *LifecycleRepository) toDomainModel(l *model.Lifecycle) (*lifecycle.Lifecycle, error) {
	lc, err := lifecycle.New(&lifecycle.LifecycleConfig{
		ProjectID:            project.ID(l.ProjectID),
		ExecutionEnvironment: l.RunsOn,
		Type:                 l.Type,
		Name:                 l.Name,
		Description:          l.Description,
		Command:              l.Command,
	})
	if err != nil {
		return nil, err
	}

	lc.SetID(lifecycle.ID(l.ID))
	lc.SetCreatedAt(l.CreatedAt)
	lc.SetUpdatedAt(l.UpdatedAt)

	return lc, nil

}

func (r *LifecycleRepository) FindByID(id int) (*lifecycle.Lifecycle, error) {
	var l model.Lifecycle
	if err := r.db.First(&l, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return r.toDomainModel(&l)
}

func (r *LifecycleRepository) Save(l *lifecycle.Lifecycle) error {
	m := &model.Lifecycle{
		ID:          l.ID().Value(),
		ProjectID:   l.ProjectID().Value(),
		Type:        l.Type(),
		Name:        l.Name().String(),
		Description: l.Description(),
		Command:     l.Command().String(),
		RunsOn:      l.ExecutionEnvironment().String(),
	}

	if !l.CreatedAt().IsZero() {
		m.CreatedAt = l.CreatedAt()
	}

	if err := r.db.Save(m).Error; err != nil {
		return err
	}

	l.SetID(lifecycle.ID(m.ID))
	l.SetCreatedAt(m.CreatedAt)
	l.SetUpdatedAt(m.UpdatedAt)

	return nil
}

func (r *LifecycleRepository) Delete(l *lifecycle.Lifecycle) error {
	m := &model.Lifecycle{
		ID: l.ID().Value(),
	}

	if err := r.db.Delete(m).Error; err != nil {
		return err
	}

	return nil
}
