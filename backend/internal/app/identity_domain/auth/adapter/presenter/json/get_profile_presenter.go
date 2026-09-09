package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter/types"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
)

type GetProfilePresenter struct{}

func NewGetProfilePresenter() types.GetProfilePresenterType {
	return &GetProfilePresenter{}
}

func (p *GetProfilePresenter) Output(output *usecase.GetProfileOutput) api.GetProfileResponseObject {
	return &api.GetProfile200JSONResponse{
		Success: true,
		Data: api.CommonAuthSpecProfile{
			Name:     output.Name,
			Username: output.Username,
			Roles:    output.Roles,
		},
	}
}

func (p *GetProfilePresenter) Error(err error) api.GetProfileResponseObject {
	switch {
	case errors.Is(err, usecase.ErrInvalidToken):
		fallthrough
	case errors.Is(err, usecase.ErrInvalidCredentials):
		fallthrough
	case errors.Is(err, usecase.ErrExpiredToken):
		return &api.GetProfile401JSONResponse{
			Success: false,
			Message: api.SharedAPIUnauthorizedResponseMessage(err.Error()),
		}
	case errors.Is(err, usecase.ErrInternalServerError):
		fallthrough
	default:
		return &api.GetProfile500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
