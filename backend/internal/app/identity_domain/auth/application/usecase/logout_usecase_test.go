package usecase_test

import (
	"errors"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"
	"judge-backend/internal/app/identity_domain/auth/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"
	"time"
)

type LogoutUseCaseMocks struct {
	refreshTokenRepository *mocks.RefreshTokenRepository
	logger                 sharedtypes.Logger
}

func NewTestLogoutUsecase() (usecase.LogoutUseCase, *LogoutUseCaseMocks) {
	refreshTokenRepository := &mocks.RefreshTokenRepository{}

	mocks := &LogoutUseCaseMocks{
		refreshTokenRepository: refreshTokenRepository,
		logger:                 logger.NewMockLogger(),
	}
	usecase := usecase.NewLogoutUseCase(mocks.refreshTokenRepository, mocks.logger)
	return usecase, mocks
}

func (mocks *LogoutUseCaseMocks) AssertExpectations(t *testing.T) {
	mocks.refreshTokenRepository.AssertExpectations(t)
}

func TestLogoutUseCase_LogoutSuccess(t *testing.T) {
	uc, mocks := NewTestLogoutUsecase()
	refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		UserID:         1,
		ExpirationTime: int64((time.Hour * 24 * 7).Seconds()),
		Token:          "refresh_token",
	})
	mocks.refreshTokenRepository.On("FindByToken", "refresh_token").Return(refreshToken, nil)
	mocks.refreshTokenRepository.On("Delete", refreshToken).Return(nil)

	input := &usecase.LogoutInput{
		RefreshToken: "refresh_token",
	}

	output, err := uc.Execute(input)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}

	mocks.AssertExpectations(t)
}

func TestLogoutUseCase_TokenNotFound(t *testing.T) {
	uc, mocks := NewTestLogoutUsecase()

	mocks.refreshTokenRepository.On("FindByToken", "invalid_token").Return(nil, usecase.ErrInvalidToken)

	input := &usecase.LogoutInput{
		RefreshToken: "invalid_token",
	}
	_, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrInvalidToken) {
		t.Errorf("expected '%v' error, got %v", usecase.ErrInvalidToken, err)
	}

	mocks.AssertExpectations(t)
}

func TestLogoutUseCase_DeleteTokenFailed(t *testing.T) {
	uc, mocks := NewTestLogoutUsecase()
	refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		UserID:         1,
		ExpirationTime: int64((time.Hour * 24 * 7).Seconds()),
		Token:          "refresh_token",
	})

	mocks.refreshTokenRepository.On("FindByToken", "refresh_token").Return(refreshToken, nil)
	mocks.refreshTokenRepository.On("Delete", refreshToken).Return(errors.New("delete failed"))

	input := &usecase.LogoutInput{
		RefreshToken: "refresh_token",
	}
	_, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrInternalServerError) {
		t.Errorf("expected 'delete failed' error, got %v", err)
	}

	mocks.AssertExpectations(t)
}
