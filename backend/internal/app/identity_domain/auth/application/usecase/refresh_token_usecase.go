package usecase

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshTokenOutput struct {
	AccessToken  Token
	RefreshToken Token
}

type RefreshTokenUseCase interface {
	Execute(input *RefreshTokenInput) (*RefreshTokenOutput, error)
}

type refreshTokenUseCase struct {
	userRepository         repository.UserRepository
	refreshTokenRepository repository.RefreshTokenRepository
	tokenService           service.TokenService
	// config is held whole rather than having the lifetime copied out at
	// construction, so a value changed at runtime — or overridden per test
	// run — takes effect on the next refresh instead of at the next restart.
	config *config.Config
	logger sharedtypes.Logger
}

func NewRefreshTokenUseCase(
	userRepository repository.UserRepository,
	refreshTokenRepository repository.RefreshTokenRepository,
	tokenService service.TokenService,
	config *config.Config,
	logger sharedtypes.Logger,
) RefreshTokenUseCase {
	return &refreshTokenUseCase{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		config:                 config,
		logger:                 logger,
	}
}

func (uc *refreshTokenUseCase) Execute(input *RefreshTokenInput) (*RefreshTokenOutput, error) {
	uc.logger.Infof("Executing RefreshTokenUseCase")

	refreshToken, err := uc.refreshTokenRepository.FindByToken(input.RefreshToken)
	if err != nil {
		uc.logger.Errorf("Error finding refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	if refreshToken == nil {
		uc.logger.Warnf("Invalid refresh token")
		return nil, ErrInvalidToken
	}

	user, err := uc.userRepository.FindByID(refreshToken.UserID().Value())
	if err != nil {
		uc.logger.Errorf("Error finding user by ID: %v", err)
		return nil, ErrInternalServerError
	}

	if user == nil {
		uc.logger.Warnf("Invalid credentials for user ID: %d", refreshToken.UserID().Value())
		return nil, ErrInvalidCredentials
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user, int64(uc.config.Auth.AccessTokenTTLSeconds))
	if err != nil {
		uc.logger.Errorf("Error generating access token: %v", err)
		return nil, ErrInternalServerError
	}

	newRefreshToken, err := uc.tokenService.GenerateRefreshToken(user, refreshToken.ExpirationTime().Value())
	if err != nil {
		uc.logger.Errorf("Error generating refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	err = uc.refreshTokenRepository.Replace(refreshToken, newRefreshToken)
	if err != nil {
		uc.logger.Errorf("Error replacing refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	uc.logger.Infof("Successfully refreshed tokens for user ID: %d", user.GetID())
	return &RefreshTokenOutput{
		AccessToken:  Token{Value: accessToken.String(), MaxAge: accessToken.MaxAge()},
		RefreshToken: Token{Value: newRefreshToken.String(), MaxAge: refreshToken.ExpirationTime().Value()},
	}, nil
}
