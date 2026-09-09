package repository

import (
	"judge-backend/internal/app/contest_domain/contest/domain/entity/project"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewProjectRepository(db *gorm.DB, logger sharedtypes.Logger) repository.ProjectRepository {
	return &ProjectRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ProjectRepository) FindByID(contestProjectID int32) (*project.Project, error) {
	r.logger.Debugf("Starting to find project by ID", "id", contestProjectID)

	cp := &model.ContestProject{}
	if err := r.db.Preload("Contest").Preload("Project").Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debugf("Project not found", "id", contestProjectID)
			return nil, httperror.New(http.StatusNotFound, "project not found")
		}

		r.logger.Errorf("Failed to find project", "id", contestProjectID, "error", err)
		return nil, err
	}

	if !cp.IsActivated() {
		r.logger.Debugf("Project is not activated", "id", contestProjectID)
		return nil, httperror.New(http.StatusNotFound, "project not found")
	}

	project, err := project.New(&project.ProjectConfig{
		ID:             cp.ID,
		Name:           cp.Project.Name,
		IsPublic:       cp.IsPublic,
		ActivateTime:   cp.ActivateTime,
		DeactivateTime: cp.DeactivateTime,
		StartTime:      cp.StartTime,
		EndTime:        cp.EndTime,
	})
	if err != nil {
		r.logger.Errorf("Failed to create project domain model", "id", contestProjectID, "error", err)
		return nil, err
	}

	r.logger.Debugf("Successfully found project", "id", contestProjectID)
	return project, nil
}
