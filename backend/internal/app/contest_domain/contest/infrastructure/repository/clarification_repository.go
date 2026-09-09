package repository

import (
	"judge-backend/internal/app/contest_domain/contest/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type ClarificationRepository struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewClarificationRepository(db *gorm.DB, logger sharedtypes.Logger) repository.ClarificationRepository {
	return &ClarificationRepository{
		db:     db,
		logger: logger,
	}
}

func intPtrToInt32(i *int) int32 {
	if i == nil {
		return 0
	}

	return int32(*i)
}

func (r *ClarificationRepository) toDomainModel(c *model.Clarification) (*clarification.Clarification, error) {
	clarification, err := clarification.New(&clarification.ClarificationConfig{
		SenderUserID:         intPtrToInt32(c.SenderUserID),
		ReceiverUserID:       intPtrToInt32(c.ReceiverUserID),
		ContestProjectID:     c.ContestProjectID,
		IsAdminBroadcast:     c.SendToAll,
		Subject:              c.Subject,
		Content:              c.Content,
		ReplyClarificationID: c.ReplyClarificationID,
	})
	if err != nil {
		return nil, err
	}

	clarification.SetID(c.ID)

	return clarification, nil
}

func (r *ClarificationRepository) FindByID(id int32) (*clarification.Clarification, error) {
	r.logger.Debugf("Starting to find clarification by ID", "id", id)

	c := &model.Clarification{}
	if err := r.db.Preload("ContestProject", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Contest")
	}).First(c, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debugf("Clarification not found", "id", id)
			return nil, nil
		}

		r.logger.Errorf("Failed to find clarification", "id", id, "error", err)
		return nil, err
	}

	if !c.ContestProject.IsActivated() {
		r.logger.Debugf("Contest project is not activated", "id", id)
		return nil, httperror.New(http.StatusNotFound, "clarification not found")
	}

	clarification, err := r.toDomainModel(c)
	if err != nil {
		r.logger.Errorf("Failed to convert to domain model", "id", id, "error", err)
		return nil, err
	}

	r.logger.Debugf("Successfully found clarification", "id", id)
	return clarification, nil
}

func (r *ClarificationRepository) Save(clarification *clarification.Clarification) error {
	r.logger.Debugf("Starting to save clarification", "id", clarification.ID())

	cp := &model.ContestProject{}
	if err := r.db.Preload("Contest").First(cp, clarification.ContestProjectID()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debugf("Contest project not found", "projectId", clarification.ContestProjectID())
			return httperror.New(http.StatusNotFound, "contest project not found")
		}

		r.logger.Errorf("Failed to find contest project", "projectId", clarification.ContestProjectID(), "error", err)
		return err
	}

	if !cp.IsActivated() {
		r.logger.Debugf("Contest project is not activated", "projectId", clarification.ContestProjectID())
		return httperror.New(http.StatusNotFound, "contest project not found")
	}

	var senderUserID int = int(clarification.SenderUserID())
	c := &model.Clarification{
		ID:                   clarification.ID(),
		SenderUserID:         &senderUserID,
		ReplyClarificationID: clarification.ReplyClarificationID(),
		ContestProjectID:     clarification.ContestProjectID(),
		Subject:              clarification.Subject().String(),
		Content:              clarification.Content(),
	}

	if err := r.db.Save(c).Error; err != nil {
		r.logger.Errorf("Failed to save clarification", "id", clarification.ID(), "error", err)
		return err
	}

	clarification.SetID(c.ID)
	r.logger.Debugf("Successfully saved clarification", "id", c.ID)

	return nil
}
