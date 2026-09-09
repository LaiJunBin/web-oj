package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/project_domain/project/domain/entity/project"
	"judge-backend/internal/app/project_domain/project/domain/repository"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(
	db *gorm.DB,
) repository.ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) toDomainModel(p *model.Project) (*project.Project, error) {
	config := &project.ProjectConfig{
		ID:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Env:          p.Env,
		TestEnv:      p.TestEnv,
		ResultParser: p.ResultParser,
	}

	if p.Image != nil {
		config.Image = &p.Image.Tag
	}

	if p.TestImage != nil {
		config.TestImage = &p.TestImage.Tag
	}

	ds := make([]*project.DynamicFieldConfig, len(p.DynamicFields))
	for i, d := range p.DynamicFields {
		ds[i] = &project.DynamicFieldConfig{
			Type:       d.Type,
			Name:       d.Name,
			Label:      d.Label,
			DefaultVal: d.DefaultValue,
			IsRequired: d.IsRequired,
		}
	}

	config.DynamicFields = ds
	project, err := project.New(config)
	if err != nil {
		return nil, err
	}

	if !p.CreatedAt.IsZero() {
		project.SetCreatedAt(p.CreatedAt)
	}

	if !p.UpdatedAt.IsZero() {
		project.SetUpdatedAt(p.UpdatedAt)
	}

	return project, nil
}

func (r *ProjectRepository) Exists(id int) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Project{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ProjectRepository) FindByID(id int) (*project.Project, error) {
	project := &model.Project{}
	if err := r.db.
		Preload("Image").
		Preload("TestImage").
		Preload("DynamicFields").
		First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return r.toDomainModel(project)
}

func (r *ProjectRepository) Save(project *project.Project) error {
	dynamicFields := make([]model.DynamicField, len(project.DynamicFields()))
	for i, d := range project.DynamicFields() {
		dynamicFields[i] = model.DynamicField{
			Type:         d.InputType().String(),
			Name:         d.Name().String(),
			Label:        d.Label().String(),
			DefaultValue: d.DefaultValue(),
			IsRequired:   d.IsRequired(),
		}
	}

	p := &model.Project{
		ID:           project.ID().Value(),
		Name:         project.Name().String(),
		Description:  project.Description().String(),
		Env:          project.EnvVars().String(),
		TestEnv:      project.TestEnvVars().String(),
		ResultParser: project.ResultParser(),
	}

	if !project.Image().IsEmpty() {
		tag := project.Image().Tag()
		p.ImageTag = &tag
	}

	if !project.TestImage().IsEmpty() {
		tag := project.TestImage().Tag()
		p.TestImageTag = &tag
	}

	if !project.CreatedAt().IsZero() {
		p.CreatedAt = project.CreatedAt()
	}

	tx := r.db.Begin()
	if err := tx.Save(p).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(p).Association("DynamicFields").Replace(&dynamicFields); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := project.SetID(p.ID); err != nil {
		return err
	}

	project.SetCreatedAt(p.CreatedAt)
	project.SetUpdatedAt(p.UpdatedAt)

	return nil
}

func (r *ProjectRepository) Delete(project *project.Project) error {
	p := &model.Project{
		ID: project.ID().Value(),
	}

	if err := r.db.Delete(p).Error; err != nil {
		return err
	}

	return nil
}
