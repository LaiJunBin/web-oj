package presenter

import (
	jsonpresenter "judge-backend/internal/app/identity_domain/auth/adapter/presenter/json"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter/types"
)

type AuthPresenter struct {
	Login        types.LoginPresenterType
	Logout       types.LogoutPresenterType
	GetProfile   types.GetProfilePresenterType
	RefreshToken types.RefreshTokenPresenterType
}

func NewAuthPresenter() *AuthPresenter {
	return &AuthPresenter{
		Login:        jsonpresenter.NewLoginPresenter(),
		Logout:       jsonpresenter.NewLogoutPresenter(),
		GetProfile:   jsonpresenter.NewGetProfilePresenter(),
		RefreshToken: jsonpresenter.NewRefreshTokenPresenter(),
	}
}
