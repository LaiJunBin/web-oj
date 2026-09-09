package application

import (
	addgroups "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/add_groups"
	addusers "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/add_users"
	listparticipants "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/list_participants"
	removegroups "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/remove_groups"
	removeusers "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/remove_users"
)

type ParticipantApplication struct {
	addUsers         addusers.Usecase
	removeUsers      removeusers.Usecase
	addGroups        addgroups.Usecase
	removeGroups     removegroups.Usecase
	listParticipants listparticipants.Usecase
}

func NewParticipantApplication(
	addUsers addusers.Usecase,
	removeUsers removeusers.Usecase,
	addGroups addgroups.Usecase,
	removeGroups removegroups.Usecase,
	listParticipants listparticipants.Usecase,
) *ParticipantApplication {
	return &ParticipantApplication{
		addUsers:         addUsers,
		removeUsers:      removeUsers,
		addGroups:        addGroups,
		removeGroups:     removeGroups,
		listParticipants: listParticipants,
	}
}

func (a *ParticipantApplication) AddUsers(input *addusers.Input) (*addusers.Output, error) {
	return a.addUsers.Execute(input)
}

func (a *ParticipantApplication) RemoveUsers(input *removeusers.Input) (*removeusers.Output, error) {
	return a.removeUsers.Execute(input)
}

func (a *ParticipantApplication) AddGroups(input *addgroups.Input) (*addgroups.Output, error) {
	return a.addGroups.Execute(input)
}

func (a *ParticipantApplication) RemoveGroups(input *removegroups.Input) (*removegroups.Output, error) {
	return a.removeGroups.Execute(input)
}

func (a *ParticipantApplication) ListParticipants(input *listparticipants.Input) (*listparticipants.Output, error) {
	return a.listParticipants.Execute(input)
}
