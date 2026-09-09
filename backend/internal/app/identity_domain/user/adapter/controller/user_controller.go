package controller

import (
	"judge-backend/api"
	sharedtypes "judge-backend/internal/app/shared/types"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter"
	"judge-backend/internal/app/identity_domain/user/application"
	createuser "judge-backend/internal/app/identity_domain/user/application/usecase/create_user"
	deleteuser "judge-backend/internal/app/identity_domain/user/application/usecase/delete_user"
	getuser "judge-backend/internal/app/identity_domain/user/application/usecase/get_user"
	importusers "judge-backend/internal/app/identity_domain/user/application/usecase/import_users"
	listrole "judge-backend/internal/app/identity_domain/user/application/usecase/list_role"
	listuser "judge-backend/internal/app/identity_domain/user/application/usecase/list_user"
	updateuser "judge-backend/internal/app/identity_domain/user/application/usecase/update_user"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	presenter *presenter.UserPresenter
	app       *application.UserApplication
	logger    sharedtypes.Logger
}

func NewUserController(
	presenter *presenter.UserPresenter,
	app *application.UserApplication,
	logger sharedtypes.Logger,
) *UserController {
	return &UserController{
		presenter: presenter,
		app:       app,
		logger:    logger,
	}
}

func (ctl *UserController) ListRole(ctx *gin.Context, request api.ListRoleRequestObject) (api.ListRoleResponseObject, error) {
	ctl.logger.Debugf("Starting to list roles")

	input := &listrole.Input{}

	output, err := ctl.app.ListRoles(input)
	if err != nil {
		ctl.logger.Errorf("Failed to list roles", "error", err)
		return ctl.presenter.ListRole.Error(err), nil
	}

	ctl.logger.Debugf("Successfully listed roles", "count", len(output.Roles))
	return ctl.presenter.ListRole.Output(output), nil
}

func (ctl *UserController) ListUser(ctx *gin.Context, request api.ListUserRequestObject) (api.ListUserResponseObject, error) {
	ctl.logger.Debugf("Starting to list users")

	input := &listuser.Input{}

	output, err := ctl.app.ListUser(input)
	if err != nil {
		ctl.logger.Errorf("Failed to list users", "error", err)
		return ctl.presenter.ListUser.Error(err), nil
	}

	ctl.logger.Debugf("Successfully listed users", "count", len(output.Users))
	return ctl.presenter.ListUser.Output(output), nil
}

func (ctl *UserController) GetUser(ctx *gin.Context, request api.GetUserRequestObject) (api.GetUserResponseObject, error) {
	ctl.logger.Debugf("Starting to get user", "userId", request.Id)

	input := &getuser.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetUser(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user", "userId", request.Id, "error", err)
		return ctl.presenter.GetUser.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user", "userId", request.Id)
	return ctl.presenter.GetUser.Output(output), nil
}

func (ctl *UserController) CreateUser(ctx *gin.Context, request api.CreateUserRequestObject) (api.CreateUserResponseObject, error) {
	ctl.logger.Debugf("Starting to create user", "username", request.Body.Username)

	input := &createuser.Input{
		Name:     request.Body.Name,
		Username: request.Body.Username,
		Password: request.Body.Password,
	}

	if request.Body.Roles != nil {
		input.Roles = *request.Body.Roles
	}

	output, err := ctl.app.CreateUser(input)
	if err != nil {
		ctl.logger.Errorf("Failed to create user", "username", request.Body.Username, "error", err)
		return ctl.presenter.CreateUser.Error(err), nil
	}

	ctl.logger.Debugf("Successfully created user", "userId", output.ID, "username", request.Body.Username)
	return ctl.presenter.CreateUser.Output(output), nil
}

func (ctl *UserController) UpdateUser(ctx *gin.Context, request api.UpdateUserRequestObject) (api.UpdateUserResponseObject, error) {
	ctl.logger.Debugf("Starting to update user", "userId", request.Id)

	input := &updateuser.Input{
		ID:       request.Id,
		Name:     request.Body.Name,
		Password: request.Body.Password,
		Roles:    request.Body.Roles,
	}

	output, err := ctl.app.UpdateUser(input)
	if err != nil {
		ctl.logger.Errorf("Failed to update user", "userId", request.Id, "error", err)
		return ctl.presenter.UpdateUser.Error(err), nil
	}

	ctl.logger.Debugf("Successfully updated user", "userId", request.Id)
	return ctl.presenter.UpdateUser.Output(output), nil
}

func (ctl *UserController) DeleteUser(ctx *gin.Context, request api.DeleteUserRequestObject) (api.DeleteUserResponseObject, error) {
	ctl.logger.Debugf("Starting to delete user", "userId", request.Id)

	input := &deleteuser.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteUser(input)
	if err != nil {
		ctl.logger.Errorf("Failed to delete user", "userId", request.Id, "error", err)
		return ctl.presenter.DeleteUser.Error(err), nil
	}

	ctl.logger.Debugf("Successfully deleted user", "userId", request.Id)
	return ctl.presenter.DeleteUser.Output(output), nil
}

type ImportUsersPayload struct {
	Users *fileheader.FileHeader `form:"users" binding:"required"`
}

func (ctl *UserController) ImportUsers(ctx *gin.Context, request api.ImportUsersRequestObject) (api.ImportUsersResponseObject, error) {
	ctl.logger.Debugf("Starting to import users")

	payload := &ImportUsersPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		ctl.logger.Errorf("Failed to bind import users payload", "error", err)
		return ctl.presenter.ImportUsers.Error(err), nil
	}

	input := &importusers.Input{
		JSONBuf: payload.Users.Buffer(),
	}

	output, err := ctl.app.ImportUsers(input)
	if err != nil {
		ctl.logger.Errorf("Failed to import users", "error", err)
		return ctl.presenter.ImportUsers.Error(err), nil
	}

	ctl.logger.Debugf("Successfully imported users", "count", output.TotalImported)
	return ctl.presenter.ImportUsers.Output(output), nil
}
