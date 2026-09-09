package repository

import (
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/activity/domain/repository"

	"gorm.io/gorm"
)

type ClarificationRepository struct {
	db *gorm.DB
}

func NewClarificationRepository(db *gorm.DB) repository.ClarificationRepository {
	return &ClarificationRepository{
		db: db,
	}
}

func int32Ptr(i *int) *int32 {
	if i == nil {
		return nil
	}

	v := int32(*i)
	return &v
}

func intPtr(i *int32) *int {
	if i == nil {
		return nil
	}

	v := int(*i)
	return &v
}

func (r *ClarificationRepository) toDomainModel(mc *model.Clarification) (*clarification.Clarification, error) {
	c, err := clarification.New(&clarification.ClarificationConfig{
		ContestProjectID: mc.ContestProjectID,
		Subject:          mc.Subject,
		Content:          mc.Content,
		SenderUserID:     int32Ptr(mc.SenderUserID),
		ReceiverUserID:   int32Ptr(mc.ReceiverUserID),
		SendToAll:        mc.SendToAll,
		IsAnswered:       mc.IsAnswered,
		ReplyToID:        mc.ReplyClarificationID,
	})
	if err != nil {
		return nil, err
	}

	c.SetID(mc.ID)
	c.SetCreatedAt(mc.CreatedAt)
	c.SetUpdatedAt(mc.UpdatedAt)

	return c, nil
}

func (r *ClarificationRepository) FindByID(id int32) (*clarification.Clarification, error) {
	mc := &model.Clarification{}
	if err := r.db.First(mc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return r.toDomainModel(mc)
}

func (r *ClarificationRepository) Save(c *clarification.Clarification) error {
	cp := &model.ContestProject{}
	if err := r.db.First(cp, c.ContestProjectID()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return repository.ErrorContestProjectNotFound
		}

		return err
	}

	if c.ReceiverUserID() != nil {
		mc := &model.User{}
		if err := r.db.First(mc, c.ReceiverUserID()).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return repository.ErrorReceiverUserNotFound
			}

			return err
		}
	}

	mc := &model.Clarification{
		ID:                   c.ID(),
		ContestProjectID:     c.ContestProjectID(),
		Subject:              c.Subject().String(),
		Content:              c.Content(),
		SenderUserID:         intPtr(c.SenderUserID()),
		ReceiverUserID:       intPtr(c.ReceiverUserID()),
		SendToAll:            c.SendToAll(),
		IsAnswered:           c.IsAnswered(),
		ReplyClarificationID: c.ReplyToID(),
	}

	mc.CreatedAt = c.CreatedAt()

	if err := r.db.Save(mc).Error; err != nil {
		return err
	}

	c.SetID(mc.ID)

	return nil
}

func (r *ClarificationRepository) Delete(c *clarification.Clarification) error {
	mc := &model.Clarification{
		ID: c.ID(),
	}

	return r.db.Delete(mc).Error
}
