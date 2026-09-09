package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter/types"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
)

type RefreshTokenPresenter struct{}

func NewRefreshTokenPresenter() types.RefreshTokenPresenterType {
	return &RefreshTokenPresenter{}
}

func (p *RefreshTokenPresenter) Output(output *usecase.RefreshTokenOutput) api.RefreshTokenResponseObject {
	// Cookie-only, like login: the rotated tokens go out as HttpOnly cookies
	// (set by the controller), never in the body where script could read them.
	return &api.RefreshToken200JSONResponse{
		Success: true,
		Data:    api.CommonAuthSpecToken{},
	}
}

func (p *RefreshTokenPresenter) Error(err error) api.RefreshTokenResponseObject {
	switch {
	case errors.Is(err, usecase.ErrInvalidToken):
		fallthrough
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return &api.RefreshToken401JSONResponse{
			Success: false,
			Message: api.SharedAPIUnauthorizedResponseMessage(err.Error()),
		}
	case errors.Is(err, usecase.ErrInternalServerError):
		fallthrough
	default:
		return &api.RefreshToken500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
