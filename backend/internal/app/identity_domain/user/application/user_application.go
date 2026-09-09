package application

import (
	createuser "judge-backend/internal/app/identity_domain/user/application/usecase/create_user"
	deleteuser "judge-backend/internal/app/identity_domain/user/application/usecase/delete_user"
	getuser "judge-backend/internal/app/identity_domain/user/application/usecase/get_user"
	importusers "judge-backend/internal/app/identity_domain/user/application/usecase/import_users"
	listrole "judge-backend/internal/app/identity_domain/user/application/usecase/list_role"
	listuser "judge-backend/internal/app/identity_domain/user/application/usecase/list_user"
	updateuser "judge-backend/internal/app/identity_domain/user/application/usecase/update_user"
)

type UserApplication struct {
	listRole    listrole.Usecase
	listUser    listuser.Usecase
	getuser     getuser.Usecase
	createUser  createuser.Usecase
	updateUser  updateuser.Usecase
	deleteUser  deleteuser.Usecase
	importUsers importusers.Usecase
}

func NewUserApplication(
	listRole listrole.Usecase,
	listUser listuser.Usecase,
	getuser getuser.Usecase,
	createUser createuser.Usecase,
	updateUser updateuser.Usecase,
	deleteUser deleteuser.Usecase,
	importUsers importusers.Usecase,
) *UserApplication {
	return &UserApplication{
		listRole:    listRole,
		listUser:    listUser,
		getuser:     getuser,
		createUser:  createUser,
		updateUser:  updateUser,
		deleteUser:  deleteUser,
		importUsers: importUsers,
	}
}

func (ua *UserApplication) ListRoles(input *listrole.Input) (*listrole.Output, error) {
	return ua.listRole.Execute(input)
}

func (ua *UserApplication) ListUser(input *listuser.Input) (*listuser.Output, error) {
	return ua.listUser.Execute(input)
}

func (ua *UserApplication) GetUser(input *getuser.Input) (*getuser.Output, error) {
	return ua.getuser.Execute(input)
}

func (ua *UserApplication) CreateUser(input *createuser.Input) (*createuser.Output, error) {
	return ua.createUser.Execute(input)
}

func (ua *UserApplication) UpdateUser(input *updateuser.Input) (*updateuser.Output, error) {
	return ua.updateUser.Execute(input)
}

func (ua *UserApplication) DeleteUser(input *deleteuser.Input) (*deleteuser.Output, error) {
	return ua.deleteUser.Execute(input)
}

func (ua *UserApplication) ImportUsers(input *importusers.Input) (*importusers.Output, error) {
	return ua.importUsers.Execute(input)
}
