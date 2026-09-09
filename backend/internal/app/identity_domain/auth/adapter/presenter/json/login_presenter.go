package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter/types"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
)

type LoginPresenter struct{}

func NewLoginPresenter() types.LoginPresenterType {
	return &LoginPresenter{}
}

func (p *LoginPresenter) Output(output *usecase.LoginOutput) api.LoginResponseObject {
	// The tokens are delivered ONLY as HttpOnly cookies (set by the controller).
	// Echoing them in the JSON body would hand a same-origin XSS the refresh
	// token in plain reach of script, defeating the point of HttpOnly. The
	// browser client is cookie-only, so the body carries no token.
	return &api.Login200JSONResponse{
		Success: true,
		Data:    api.CommonAuthSpecToken{},
	}
}

func (p *LoginPresenter) Error(err error) api.LoginResponseObject {
	switch {
	case errors.Is(err, usecase.ErrBadRequest):
		return &api.Login400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return &api.Login401JSONResponse{
			Success: false,
			Message: api.SharedAPIUnauthorizedResponseMessage(err.Error()),
		}
	case errors.Is(err, usecase.ErrInternalServerError):
		fallthrough
	default:
		return &api.Login500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
