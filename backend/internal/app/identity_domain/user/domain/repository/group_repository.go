package repository

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/group"
)

type GroupRepository interface {
	Save(group *group.Group) error
	GetByID(id int) (*group.Group, error)
	FindAll() ([]*group.Group, error)
	Delete(id int) error
	AddUsersToGroup(groupID int, userIDs []int) error
	RemoveUsersFromGroup(groupID int, userIDs []int) error
	GetUserIDs(groupID int) ([]int, error)
}
