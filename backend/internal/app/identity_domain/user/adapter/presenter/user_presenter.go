package presenter

import (
	jsonpresenter "judge-backend/internal/app/identity_domain/user/adapter/presenter/json"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
)

type UserPresenter struct {
	ListRole    types.ListRolePresenter
	ListUser    types.ListUserPresenter
	GetUser     types.GetUserPresenter
	CreateUser  types.CreateUserPresenter
	UpdateUser  types.UpdateUserPresenter
	DeleteUser  types.DeleteUserPresenter
	ImportUsers types.ImportUsersPresenter
}

func NewUserPresenter() *UserPresenter {
	return &UserPresenter{
		ListRole:    jsonpresenter.NewListRolePresenter(),
		ListUser:    jsonpresenter.NewListUserPresenter(),
		GetUser:     jsonpresenter.NewGetUserPresenter(),
		CreateUser:  jsonpresenter.NewCreateUserPresenter(),
		UpdateUser:  jsonpresenter.NewUpdateUserPresenter(),
		DeleteUser:  jsonpresenter.NewDeleteUserPresenter(),
		ImportUsers: jsonpresenter.NewImportUsersPresenter(),
	}
}
