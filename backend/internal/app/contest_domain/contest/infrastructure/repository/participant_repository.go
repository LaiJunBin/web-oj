package repository

import (
	"judge-backend/internal/app/contest_domain/contest/domain/repository"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ParticipantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) repository.ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) AddUsers(contestID int32, userIDs []int) error {
	rows := make([]model.ContestParticipant, 0, len(userIDs))
	for _, uid := range userIDs {
		id := uid
		rows = append(rows, model.ContestParticipant{ContestID: contestID, UserID: &id})
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (r *ParticipantRepository) RemoveUsers(contestID int32, userIDs []int) error {
	return r.db.Where("contest_id = ? AND user_id IN ?", contestID, userIDs).Delete(&model.ContestParticipant{}).Error
}

func (r *ParticipantRepository) AddGroups(contestID int32, groupIDs []int) error {
	rows := make([]model.ContestParticipant, 0, len(groupIDs))
	for _, gid := range groupIDs {
		id := gid
		rows = append(rows, model.ContestParticipant{ContestID: contestID, GroupID: &id})
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (r *ParticipantRepository) RemoveGroups(contestID int32, groupIDs []int) error {
	return r.db.Where("contest_id = ? AND group_id IN ?", contestID, groupIDs).Delete(&model.ContestParticipant{}).Error
}

func (r *ParticipantRepository) GetUserIDs(contestID int32) ([]int, error) {
	var rows []model.ContestParticipant
	if err := r.db.Where("contest_id = ? AND user_id IS NOT NULL", contestID).Order("user_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row.UserID != nil {
			ids = append(ids, *row.UserID)
		}
	}
	return ids, nil
}

func (r *ParticipantRepository) GetGroupIDs(contestID int32) ([]int, error) {
	var rows []model.ContestParticipant
	if err := r.db.Where("contest_id = ? AND group_id IS NOT NULL", contestID).Order("group_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows))
	for _, row := range rows {
		if row.GroupID != nil {
			ids = append(ids, *row.GroupID)
		}
	}
	return ids, nil
}
