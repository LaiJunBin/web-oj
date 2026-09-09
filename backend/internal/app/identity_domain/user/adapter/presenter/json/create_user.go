package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	createuser "judge-backend/internal/app/identity_domain/user/application/usecase/create_user"
)

type CreateUserPresenter struct{}

func NewCreateUserPresenter() types.CreateUserPresenter {
	return &CreateUserPresenter{}
}

func (p *CreateUserPresenter) Output(output *createuser.Output) api.CreateUserResponseObject {
	return &api.CreateUser200JSONResponse{
		Success: true,
		Data:    int32(output.ID),
	}
}

func (p *CreateUserPresenter) Error(err error) api.CreateUserResponseObject {
	switch {
	case errors.Is(err, createuser.ErrorNameIsRequired):
		fallthrough
	case errors.Is(err, createuser.ErrorUsernameIsRequired):
		fallthrough
	case errors.Is(err, createuser.ErrorUsernameReserved):
		fallthrough
	case errors.Is(err, createuser.ErrorPasswordIsRequired):
		fallthrough
	case errors.Is(err, createuser.ErrorInvalidRoles):
		fallthrough
	case errors.Is(err, createuser.ErrorDuplicateUsername):
		return &api.CreateUser400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, createuser.ErrorInternalServer):
		fallthrough
	default:
		return &api.CreateUser500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
