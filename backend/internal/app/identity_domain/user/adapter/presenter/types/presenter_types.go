package types

import (
	"judge-backend/api"
	createuser "judge-backend/internal/app/identity_domain/user/application/usecase/create_user"
	deleteuser "judge-backend/internal/app/identity_domain/user/application/usecase/delete_user"
	getuser "judge-backend/internal/app/identity_domain/user/application/usecase/get_user"
	importusers "judge-backend/internal/app/identity_domain/user/application/usecase/import_users"
	listrole "judge-backend/internal/app/identity_domain/user/application/usecase/list_role"
	listuser "judge-backend/internal/app/identity_domain/user/application/usecase/list_user"
	updateuser "judge-backend/internal/app/identity_domain/user/application/usecase/update_user"
)

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

type ListRolePresenter Presenter[listrole.Output, api.ListRoleResponseObject]
type ListUserPresenter Presenter[listuser.Output, api.ListUserResponseObject]
type GetUserPresenter Presenter[getuser.Output, api.GetUserResponseObject]
type CreateUserPresenter Presenter[createuser.Output, api.CreateUserResponseObject]
type UpdateUserPresenter Presenter[updateuser.Output, api.UpdateUserResponseObject]
type DeleteUserPresenter Presenter[deleteuser.Output, api.DeleteUserResponseObject]
type ImportUsersPresenter Presenter[importusers.Output, api.ImportUsersResponseObject]
