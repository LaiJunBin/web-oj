package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter/types"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
)

type LogoutPresenter struct{}

func NewLogoutPresenter() types.LogoutPresenterType {
	return &LogoutPresenter{}
}

func (p *LogoutPresenter) Output(output *usecase.LogoutOutput) api.LogoutResponseObject {
	return &api.Logout200JSONResponse{
		Success: true,
	}
}

func (p *LogoutPresenter) Error(err error) api.LogoutResponseObject {
	switch {
	case errors.Is(err, usecase.ErrInvalidToken):
		return &api.Logout401JSONResponse{
			Success: false,
			Message: api.SharedAPIUnauthorizedResponseMessage(err.Error()),
		}
	case errors.Is(err, usecase.ErrInternalServerError):
		fallthrough
	default:
		return &api.Logout500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
