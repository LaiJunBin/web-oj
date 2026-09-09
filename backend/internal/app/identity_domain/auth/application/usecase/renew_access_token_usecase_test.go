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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type RenewAccessTokenUseCaseMocks struct {
	userRepository         *mocks.UserRepository
	refreshTokenRepository *mocks.RefreshTokenRepository
	tokenService           *mocks.TokenService
	logger                 sharedtypes.Logger
}

func NewTestRenewAccessTokenUseCase() (usecase.RenewAccessTokenUseCase, *RenewAccessTokenUseCaseMocks) {
	m := &RenewAccessTokenUseCaseMocks{
		userRepository:         &mocks.UserRepository{},
		refreshTokenRepository: &mocks.RefreshTokenRepository{},
		tokenService:           &mocks.TokenService{},
		logger:                 logger.NewMockLogger(),
	}
	uc := usecase.NewRenewAccessTokenUseCase(m.userRepository, m.refreshTokenRepository, m.tokenService, testAuthConfig(), m.logger)
	return uc, m
}

func (m *RenewAccessTokenUseCaseMocks) AssertExpectations(t *testing.T) {
	m.userRepository.AssertExpectations(t)
	m.refreshTokenRepository.AssertExpectations(t)
	m.tokenService.AssertExpectations(t)
}

func renewTestUser() *user.User {
	u := user.MustNew("User", "user01", "password")
	u.SetID(1)
	u.SetRoles(role.RoleAdmin)
	return u
}

// This is the whole point of the usecase. The middleware used to renew via
// RefreshTokenUseCase, which rotates: it mints a new refresh token and
// Replace()s the old one, and Replace is a Redis RENAME that destroys the
// old key. Contest pages hold three SSE connections open at once, all
// carrying the same refresh cookie, so when the access token expired all
// three renewed simultaneously — the first RENAME won and the other two
// found their token gone, returned 401, and logged the contestant out. On
// top of that the rotated cookie never reached the browser, because it was
// issued on an SSE response behind a proxy.
func TestRenewAccessTokenUsecase_DoesNotRotateTheRefreshToken(t *testing.T) {
	uc, m := NewTestRenewAccessTokenUseCase()
	mockUser := renewTestUser()
	mockRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "refresh_token",
		UserID:         mockUser.GetID(),
		ExpirationTime: int64(time.Minute.Seconds()),
	})
	newAccessToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "encrypted_" + strconv.Itoa(mockUser.GetID()),
		Roles:           mockUser.GetRoles(),
		ExpirationTime:  testAccessTTL,
	})

	m.refreshTokenRepository.On("FindByToken", "refresh_token").Return(mockRefreshToken, nil)
	m.userRepository.On("FindByID", mockUser.GetID()).Return(mockUser, nil)
	m.tokenService.On("GenerateAccessToken", mockUser, int64(testAccessTTL)).Return(newAccessToken, nil)

	output, err := uc.Execute("refresh_token")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if output.AccessToken.Value != newAccessToken.String() {
		t.Errorf("expected access token %v, got %v", newAccessToken.String(), output.AccessToken.Value)
	}

	// Neither of these may ever be reached: minting a refresh token is
	// harmless on its own, but Replace is what destroys the old key.
	m.tokenService.AssertNotCalled(t, "GenerateRefreshToken", mock.Anything, mock.Anything)
	m.refreshTokenRepository.AssertNotCalled(t, "Replace", mock.Anything, mock.Anything)
	m.refreshTokenRepository.AssertNotCalled(t, "Delete", mock.Anything)
	m.refreshTokenRepository.AssertNotCalled(t, "Save", mock.Anything)

	m.AssertExpectations(t)
}

// The access lifetime comes from config, not from whatever the refresh
// token's remaining TTL happens to be.
func TestRenewAccessTokenUsecase_UsesTheConfiguredAccessLifetime(t *testing.T) {
	uc, m := NewTestRenewAccessTokenUseCase()
	mockUser := renewTestUser()
	mockRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "refresh_token",
		UserID:         mockUser.GetID(),
		ExpirationTime: 999999,
	})
	newAccessToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "encrypted_1",
		Roles:           mockUser.GetRoles(),
		ExpirationTime:  testAccessTTL,
	})

	m.refreshTokenRepository.On("FindByToken", "refresh_token").Return(mockRefreshToken, nil)
	m.userRepository.On("FindByID", mockUser.GetID()).Return(mockUser, nil)
	m.tokenService.On("GenerateAccessToken", mockUser, int64(testAccessTTL)).Return(newAccessToken, nil)

	if _, err := uc.Execute("refresh_token"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	m.AssertExpectations(t)
}

func TestRenewAccessTokenUsecase_TokenNotFound(t *testing.T) {
	uc, m := NewTestRenewAccessTokenUseCase()

	m.refreshTokenRepository.On("FindByToken", "invalid_token").Return(nil, nil)

	_, err := uc.Execute("invalid_token")
	if !errors.Is(err, usecase.ErrInvalidToken) {
		t.Errorf("expected '%v', got %v", usecase.ErrInvalidToken, err)
	}

	m.AssertExpectations(t)
}

func TestRenewAccessTokenUsecase_TokenLookupFails(t *testing.T) {
	uc, m := NewTestRenewAccessTokenUseCase()

	m.refreshTokenRepository.On("FindByToken", "refresh_token").Return(nil, errors.New("redis down"))

	_, err := uc.Execute("refresh_token")
	if !errors.Is(err, usecase.ErrInternalServerError) {
		t.Errorf("expected '%v', got %v", usecase.ErrInternalServerError, err)
	}

	m.AssertExpectations(t)
}

func TestRenewAccessTokenUsecase_UserNotFound(t *testing.T) {
	uc, m := NewTestRenewAccessTokenUseCase()
	mockRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "refresh_token",
		UserID:         1,
		ExpirationTime: int64(time.Minute.Seconds()),
	})

	m.refreshTokenRepository.On("FindByToken", "refresh_token").Return(mockRefreshToken, nil)
	m.userRepository.On("FindByID", 1).Return(nil, nil)

	_, err := uc.Execute("refresh_token")
	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Errorf("expected '%v', got %v", usecase.ErrInvalidCredentials, err)
	}

	m.AssertExpectations(t)
}

func TestRenewAccessTokenUsecase_GenerateFails(t *testing.T) {
	uc, m := NewTestRenewAccessTokenUseCase()
	mockUser := renewTestUser()
	mockRefreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          "refresh_token",
		UserID:         mockUser.GetID(),
		ExpirationTime: int64(time.Minute.Seconds()),
	})

	m.refreshTokenRepository.On("FindByToken", "refresh_token").Return(mockRefreshToken, nil)
	m.userRepository.On("FindByID", mockUser.GetID()).Return(mockUser, nil)
	m.tokenService.On("GenerateAccessToken", mockUser, int64(testAccessTTL)).Return(nil, errors.New("boom"))

	_, err := uc.Execute("refresh_token")
	if !errors.Is(err, usecase.ErrInternalServerError) {
		t.Errorf("expected '%v', got %v", usecase.ErrInternalServerError, err)
	}

	m.AssertExpectations(t)
}
