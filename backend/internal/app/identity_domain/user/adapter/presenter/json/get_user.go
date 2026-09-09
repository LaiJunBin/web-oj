package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	getuser "judge-backend/internal/app/identity_domain/user/application/usecase/get_user"
)

type GetUserPresenter struct{}

func NewGetUserPresenter() types.GetUserPresenter {
	return &GetUserPresenter{}
}

func (p *GetUserPresenter) Output(output *getuser.Output) api.GetUserResponseObject {
	return &api.GetUser200JSONResponse{
		Success: true,
		Data: api.AdminUserSpecUser{
			Id:        output.User.ID,
			Username:  output.User.Username,
			Name:      output.User.Name,
			Roles:     output.User.Roles,
			CreatedAt: output.User.CreatedAt,
			UpdatedAt: output.User.UpdatedAt,
		},
	}
}

func (p *GetUserPresenter) Error(err error) api.GetUserResponseObject {
	switch {
	case errors.Is(err, getuser.ErrorInvalidID):
		return &api.GetUser400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getuser.ErrorUserNotFound):
		return &api.GetUser404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, getuser.ErrorInternalServer):
		fallthrough
	default:
		return &api.GetUser500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
