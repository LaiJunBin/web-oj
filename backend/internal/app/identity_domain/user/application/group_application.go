package application

import (
	adduserstogroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/add_users_to_group"
	creategroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/create_group"
	deletegroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/delete_group"
	getgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/get_group"
	listgroups "judge-backend/internal/app/identity_domain/user/application/usecase/group/list_groups"
	removeusersfromgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/remove_users_from_group"
	updategroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/update_group"
)

type GroupApplication struct {
	createGroup          creategroup.Usecase
	getGroup             getgroup.Usecase
	listGroups           listgroups.Usecase
	updateGroup          updategroup.Usecase
	deleteGroup          deletegroup.Usecase
	addUsersToGroup      adduserstogroup.Usecase
	removeUsersFromGroup removeusersfromgroup.Usecase
}

func NewGroupApplication(
	createGroup creategroup.Usecase,
	getGroup getgroup.Usecase,
	listGroups listgroups.Usecase,
	updateGroup updategroup.Usecase,
	deleteGroup deletegroup.Usecase,
	addUsersToGroup adduserstogroup.Usecase,
	removeUsersFromGroup removeusersfromgroup.Usecase,
) *GroupApplication {
	return &GroupApplication{
		createGroup:          createGroup,
		getGroup:             getGroup,
		listGroups:           listGroups,
		updateGroup:          updateGroup,
		deleteGroup:          deleteGroup,
		addUsersToGroup:      addUsersToGroup,
		removeUsersFromGroup: removeUsersFromGroup,
	}
}

func (ga *GroupApplication) CreateGroup(input *creategroup.Input) (*creategroup.Output, error) {
	return ga.createGroup.Execute(input)
}

func (ga *GroupApplication) GetGroup(input *getgroup.Input) (*getgroup.Output, error) {
	return ga.getGroup.Execute(input)
}

func (ga *GroupApplication) ListGroups(input *listgroups.Input) (*listgroups.Output, error) {
	return ga.listGroups.Execute(input)
}

func (ga *GroupApplication) UpdateGroup(input *updategroup.Input) (*updategroup.Output, error) {
	return ga.updateGroup.Execute(input)
}

func (ga *GroupApplication) DeleteGroup(input *deletegroup.Input) (*deletegroup.Output, error) {
	return ga.deleteGroup.Execute(input)
}

func (ga *GroupApplication) AddUsersToGroup(input *adduserstogroup.Input) (*adduserstogroup.Output, error) {
	return ga.addUsersToGroup.Execute(input)
}

func (ga *GroupApplication) RemoveUsersFromGroup(input *removeusersfromgroup.Input) (*removeusersfromgroup.Output, error) {
	return ga.removeUsersFromGroup.Execute(input)
}
