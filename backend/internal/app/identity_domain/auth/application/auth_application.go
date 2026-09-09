package application

import "judge-backend/internal/app/identity_domain/auth/application/usecase"

type AuthApplication struct {
	login        usecase.LoginUseCase
	logout       usecase.LogoutUseCase
	getProfile   usecase.GetProfileUseCase
	refreshToken usecase.RefreshTokenUseCase
}

func NewAuthApplication(
	login usecase.LoginUseCase,
	logout usecase.LogoutUseCase,
	getProfile usecase.GetProfileUseCase,
	refreshToken usecase.RefreshTokenUseCase,
) *AuthApplication {
	return &AuthApplication{
		login:        login,
		logout:       logout,
		getProfile:   getProfile,
		refreshToken: refreshToken,
	}
}

func (s *AuthApplication) Login(input *usecase.LoginInput) (*usecase.LoginOutput, error) {
	return s.login.Execute(input)
}

func (s *AuthApplication) Logout(input *usecase.LogoutInput) (*usecase.LogoutOutput, error) {
	return s.logout.Execute(input)
}

func (s *AuthApplication) GetProfile(input *usecase.GetProfileInput) (*usecase.GetProfileOutput, error) {
	return s.getProfile.Execute(input)
}

func (s *AuthApplication) RefreshToken(input *usecase.RefreshTokenInput) (*usecase.RefreshTokenOutput, error) {
	return s.refreshToken.Execute(input)
}
