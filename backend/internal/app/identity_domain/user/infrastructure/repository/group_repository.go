package repository

import (
	"errors"
	"judge-backend/internal/app/identity_domain/user/domain/entity/group"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) repository.GroupRepository {
	return &GroupRepository{
		db: db,
	}
}

func (r *GroupRepository) toDomainGroup(g *model.Group) (*group.Group, error) {
	domainGroup, err := group.New(&group.GroupConfig{
		Name:        g.Name,
		Description: g.Description,
	})
	if err != nil {
		return nil, err
	}
	domainGroup.SetID(g.ID)
	return domainGroup, nil
}

func (r *GroupRepository) toModelGroup(g *group.Group) (*model.Group, error) {
	modelGroup := &model.Group{}
	if err := r.db.First(modelGroup, g.GetID()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return modelGroup, nil
		}
		return nil, err
	}
	return modelGroup, nil
}

func (r *GroupRepository) Save(group *group.Group) error {
	g, err := r.toModelGroup(group)
	if err != nil {
		return err
	}

	g.Name = group.GetName()
	g.Description = group.GetDescription()

	if err := r.db.Save(g).Error; err != nil {
		return err
	}

	group.SetID(g.ID)
	return nil
}

func (r *GroupRepository) GetByID(id int) (*group.Group, error) {
	g := &model.Group{}
	if err := r.db.First(g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomainGroup(g)
}

func (r *GroupRepository) FindAll() ([]*group.Group, error) {
	var models []model.Group
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}

	var groups []*group.Group
	for _, m := range models {
		g, err := r.toDomainGroup(&m)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *GroupRepository) Delete(id int) error {
	return r.db.Delete(&model.Group{}, id).Error
}

func (r *GroupRepository) AddUsersToGroup(groupID int, userIDs []int) error {
	if len(userIDs) == 0 {
		return nil
	}

	var userGroups []model.UserGroup
	for _, uid := range userIDs {
		userGroups = append(userGroups, model.UserGroup{
			UserID:  uid,
			GroupID: groupID,
		})
	}
	// On conflict, do nothing (ignore duplicate inserts)
	return r.db.Clauses(
		clause.OnConflict{DoNothing: true},
	).Create(&userGroups).Error
}

func (r *GroupRepository) RemoveUsersFromGroup(groupID int, userIDs []int) error {
	return r.db.Where("group_id = ? AND user_id IN ?", groupID, userIDs).Delete(&model.UserGroup{}).Error
}

func (r *GroupRepository) GetUserIDs(groupID int) ([]int, error) {
	var userGroups []model.UserGroup
	if err := r.db.Where("group_id = ?", groupID).Order("user_id").Find(&userGroups).Error; err != nil {
		return nil, err
	}

	userIDs := make([]int, 0, len(userGroups))
	for _, ug := range userGroups {
		userIDs = append(userIDs, ug.UserID)
	}
	return userIDs, nil
}
