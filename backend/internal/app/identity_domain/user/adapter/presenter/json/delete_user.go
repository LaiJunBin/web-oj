package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	deleteuser "judge-backend/internal/app/identity_domain/user/application/usecase/delete_user"
)

type DeleteUserPresenter struct{}

func NewDeleteUserPresenter() types.DeleteUserPresenter {
	return &DeleteUserPresenter{}
}

func (p *DeleteUserPresenter) Output(output *deleteuser.Output) api.DeleteUserResponseObject {
	return &api.DeleteUser200JSONResponse{
		Success: true,
	}
}

func (p *DeleteUserPresenter) Error(err error) api.DeleteUserResponseObject {
	switch {
	case errors.Is(err, deleteuser.ErrorInvalidInput):
		return &api.DeleteUser400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteuser.ErrorUserNotFound):
		return &api.DeleteUser404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteuser.ErrorInternalServer):
		fallthrough
	default:
		return &api.DeleteUser500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
