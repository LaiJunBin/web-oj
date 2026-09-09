package usecase_test

import (
	"errors"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	"judge-backend/internal/app/identity_domain/auth/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
	"strconv"
	"testing"
	"time"
)

// testAccessTTL is deliberately not 3600. Token lifetimes used to be
// hardcoded to time.Hour, and an expectation of 3600 would have passed
// against that hardcoding — these odd numbers only match if the value
// genuinely came from config.
const (
	testAccessTTL  = 111
	testRefreshTTL = 222
)

func testAuthConfig() *config.Config {
	return &config.Config{
		Auth: &config.AuthConfig{
			AccessTokenTTLSeconds:  testAccessTTL,
			RefreshTokenTTLSeconds: testRefreshTTL,
		},
	}
}

type RefreshTokenUseCaseMocks struct {
	userRepository         *mocks.UserRepository
	refreshTokenRepository *mocks.RefreshTokenRepository
	tokenService           *mocks.TokenService
	logger                 sharedtypes.Logger
}

func NewTestRefreshTokenUseCase() (usecase.RefreshTokenUseCase, *RefreshTokenUseCaseMocks) {
	userRepository := &mocks.UserRepository{}
	refreshTokenRepository := &mocks.RefreshTokenRepository{}
	tokenService := &mocks.TokenService{}

	mocks := &RefreshTokenUseCaseMocks{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		logger:                 logger.NewMockLogger(),
	}
	usecase := usecase.NewRefreshTokenUseCase(mocks.userRepository, mocks.refreshTokenRepository, mocks.tokenService, testAuthConfig(), mocks.logger)
	return usecase, mocks
}

func (mocks *RefreshTokenUseCaseMocks) AssertExpectations(t *testing.T) {
	mocks.userRepository.AssertExpectations(t)
	mocks.refreshTokenRepository.AssertExpectations(t)
	mocks.tokenService.AssertExpectations(t)
}

func TestRefreshTokenUsecase_Success(t *testing.T) {
	uc, mocks := NewTestRefreshTokenUseCase()
	mockUser := user.MustNew("User", "user01", "password")
	mockUser.SetID(1)
	mockUser.SetRoles(role.RoleAdmin)
	mockRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "refresh_token",
		UserID:         mockUser.GetID(),
		ExpirationTime: int64(time.Minute.Seconds()),
	})
	mockNewJwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "encrypted_" + strconv.Itoa(mockUser.GetID()),
		Roles:           mockUser.GetRoles(),
		ExpirationTime:  int64(time.Hour.Seconds()),
	})
	mockNewRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "new_refresh_token",
		UserID:         mockUser.GetID(),
		ExpirationTime: int64(time.Hour.Seconds()),
	})
	mocks.refreshTokenRepository.On("FindByToken", "refresh_token").Return(mockRefreshToken, nil)
	mocks.userRepository.On("FindByID", mockUser.GetID()).Return(mockUser, nil)
	mocks.tokenService.On("GenerateAccessToken", mockUser, int64(testAccessTTL)).Return(mockNewJwtToken, nil)
	mocks.tokenService.On("GenerateRefreshToken", mockUser, mockRefreshToken.ExpirationTime().Value()).Return(mockNewRefreshToken, nil)
	mocks.refreshTokenRepository.On("Replace", mockRefreshToken, mockNewRefreshToken).Return(nil)

	input := &usecase.RefreshTokenInput{
		RefreshToken: "refresh_token",
	}
	output, err := uc.Execute(input)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if output == nil {
		t.Errorf("expected not nil, got nil")
		return
	}

	if output.AccessToken.Value != mockNewJwtToken.String() {
		t.Errorf("expected token to be %v, got %v", mockNewJwtToken.String(), output.AccessToken)
	}

	if output.RefreshToken.Value != mockNewRefreshToken.String() {
		t.Errorf("expected refresh token to be %v, got %v", mockNewRefreshToken.String(), output.RefreshToken)
	}

	mocks.AssertExpectations(t)
}

func TestRefreshTokenUsecase_TokenNotFound(t *testing.T) {
	uc, mocks := NewTestRefreshTokenUseCase()

	mocks.refreshTokenRepository.On("FindByToken", "invalid_token").Return(nil, nil)

	input := &usecase.RefreshTokenInput{
		RefreshToken: "invalid_token",
	}
	_, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrInvalidToken) {
		t.Errorf("expected '%v' error, got %v", usecase.ErrInvalidToken, err)
	}

	mocks.AssertExpectations(t)
}

func TestRefreshTokenUsecase_UserNotFound(t *testing.T) {
	uc, mocks := NewTestRefreshTokenUseCase()
	mockRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "refresh_token",
		UserID:         1,
		ExpirationTime: int64(time.Minute.Seconds()),
	})

	mocks.refreshTokenRepository.On("FindByToken", "refresh_token").Return(mockRefreshToken, nil)
	mocks.userRepository.On("FindByID", 1).Return(nil, nil)

	input := &usecase.RefreshTokenInput{
		RefreshToken: "refresh_token",
	}
	_, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Errorf("expected '%v' error, got %v", usecase.ErrInvalidCredentials, err)
	}

	mocks.AssertExpectations(t)
}
