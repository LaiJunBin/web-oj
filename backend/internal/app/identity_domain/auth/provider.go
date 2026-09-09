package auth

import (
	"judge-backend/internal/app/identity_domain/auth/adapter/controller"
	"judge-backend/internal/app/identity_domain/auth/adapter/cookie"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter"
	"judge-backend/internal/app/identity_domain/auth/application"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/repository"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/service"
	"judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *provider.Provider {
	return provider.New(provider.Config{
		Infrastructures: []interface{}{
			repository.NewUserRepository,
			repository.NewRefreshTokenRepository,
			service.NewTokenService,
			service.NewHashService,
			service.NewEncryptService,
			service.NewRandomService,
			service.NewUserRoleService,
		},
		Controllers: []interface{}{
			controller.NewAuthController,
		},
		Presenters: []interface{}{
			presenter.NewAuthPresenter,
		},
		Usecases: []interface{}{
			usecase.NewLoginUseCase,
			usecase.NewLogoutUseCase,
			usecase.NewGetProfileUseCase,
			usecase.NewRefreshTokenUseCase,
			usecase.NewRenewAccessTokenUseCase,
		},
		Applications: []interface{}{
			application.NewAuthApplication,
		},
		Providers: []interface{}{
			// Shared by the controller and the authorization middleware, so
			// both write and clear the auth cookies with identical
			// attributes — a mismatch makes a logout silently not log out.
			cookie.NewCodec,
		},
	})
}
