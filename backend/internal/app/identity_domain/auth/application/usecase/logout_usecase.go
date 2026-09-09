package usecase

import (
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutOutput struct{}

type LogoutUseCase interface {
	Execute(input *LogoutInput) (*LogoutOutput, error)
}

type logoutUseCase struct {
	refreshTokenRepository repository.RefreshTokenRepository
	logger                 sharedtypes.Logger
}

func NewLogoutUseCase(
	refreshTokenRepository repository.RefreshTokenRepository,
	logger sharedtypes.Logger,
) LogoutUseCase {
	return &logoutUseCase{
		refreshTokenRepository: refreshTokenRepository,
		logger:                 logger,
	}
}

func (uc *logoutUseCase) Execute(input *LogoutInput) (*LogoutOutput, error) {
	uc.logger.Infof("Executing LogoutUseCase")

	refreshToken, err := uc.refreshTokenRepository.FindByToken(input.RefreshToken)
	if err != nil {
		uc.logger.Errorf("Error finding refresh token: %v", err)
		return nil, ErrInvalidToken
	}

	if refreshToken == nil {
		uc.logger.Warnf("Invalid refresh token")
		return nil, ErrInvalidToken
	}

	err = uc.refreshTokenRepository.Delete(refreshToken)
	if err != nil {
		uc.logger.Errorf("Error deleting refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	uc.logger.Infof("Successfully logged out user")
	return nil, nil
}
