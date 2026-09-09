package types

import (
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
)

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

type LoginPresenterType Presenter[usecase.LoginOutput, api.LoginResponseObject]
type LogoutPresenterType Presenter[usecase.LogoutOutput, api.LogoutResponseObject]
type GetProfilePresenterType Presenter[usecase.GetProfileOutput, api.GetProfileResponseObject]
type RefreshTokenPresenterType Presenter[usecase.RefreshTokenOutput, api.RefreshTokenResponseObject]
