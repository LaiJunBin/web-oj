package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	updateuser "judge-backend/internal/app/identity_domain/user/application/usecase/update_user"
)

type UpdateUserPresenter struct{}

func NewUpdateUserPresenter() types.UpdateUserPresenter {
	return &UpdateUserPresenter{}
}

func (p *UpdateUserPresenter) Output(output *updateuser.Output) api.UpdateUserResponseObject {
	return &api.UpdateUser200JSONResponse{
		Success: true,
	}
}

func (p *UpdateUserPresenter) Error(err error) api.UpdateUserResponseObject {
	switch {
	case errors.Is(err, updateuser.ErrorInvalidInput):
		fallthrough
	case errors.Is(err, updateuser.ErrorInvalidRole):
		return &api.UpdateUser400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, updateuser.ErrorUserNotFound):
		return &api.UpdateUser404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updateuser.ErrorInternalServer):
		fallthrough
	default:
		return &api.UpdateUser500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
