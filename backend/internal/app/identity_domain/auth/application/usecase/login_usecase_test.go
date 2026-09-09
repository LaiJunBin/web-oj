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
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type LoginUseCaseMocks struct {
	userRepository         *mocks.UserRepository
	refreshTokenRepository *mocks.RefreshTokenRepository
	tokenService           *mocks.TokenService
	hashService            *mocks.HashService
	logger                 sharedtypes.Logger
}

func NewTestLoginUsecase() (usecase.LoginUseCase, *LoginUseCaseMocks) {
	userRepository := &mocks.UserRepository{}
	refreshTokenRepository := &mocks.RefreshTokenRepository{}
	tokenService := &mocks.TokenService{}
	hashService := &mocks.HashService{}

	mocks := &LoginUseCaseMocks{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		hashService:            hashService,
		logger:                 logger.NewMockLogger(),
	}
	usecase := usecase.NewLoginUseCase(mocks.userRepository, mocks.refreshTokenRepository, mocks.tokenService, mocks.hashService, testAuthConfig(), mocks.logger)
	return usecase, mocks
}

func (mocks *LoginUseCaseMocks) AssertExpectations(t *testing.T) {
	mocks.userRepository.AssertExpectations(t)
	mocks.refreshTokenRepository.AssertExpectations(t)
	mocks.tokenService.AssertExpectations(t)
	mocks.hashService.AssertExpectations(t)
}

func TestLoginUseCase_LoginSuccess(t *testing.T) {
	uc, mocks := NewTestLoginUsecase()
	mockUser := user.MustNew("User", "user01", "password")
	mockUser.SetID(1)
	mockUser.SetRoles(role.RoleAdmin)
	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "user01",
		Roles:           mockUser.GetRoles(),
		ExpirationTime:  int64(time.Hour.Seconds()),
	})
	refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		UserID:         mockUser.GetID(),
		ExpirationTime: int64((time.Hour * 24 * 7).Seconds()),
		Token:          "refresh_token",
	})

	mocks.userRepository.On("FindByUsername", "user01").Return(mockUser, nil)
	mocks.hashService.On("Compare", "1234", "password").Return(true)
	mocks.tokenService.On("GenerateAccessToken", mockUser, int64(testAccessTTL)).Return(jwtToken, nil)
	mocks.tokenService.On("GenerateRefreshToken", mockUser, int64(testRefreshTTL)).Return(refreshToken, nil)
	mocks.refreshTokenRepository.On("Save", refreshToken).Return(nil)

	input := &usecase.LoginInput{
		Username: "user01",
		Password: "1234",
	}

	output, err := uc.Execute(input)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if output.AccessToken.Value != jwtToken.String() {
		t.Errorf("expected token, got %v", output.AccessToken)
	}

	if output.RefreshToken.Value != "refresh_token" {
		t.Errorf("expected refresh token, got %v", output.RefreshToken)
	}

	mocks.AssertExpectations(t)
}

func TestLoginUseCase_EmptyUsernameAndPassword(t *testing.T) {
	uc, _ := NewTestLoginUsecase()
	input := &usecase.LoginInput{
		Username: "",
		Password: "",
	}

	output, err := uc.Execute(input)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if !errors.Is(err, usecase.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}
}

func TestLoginUseCase_InvalidCredentials(t *testing.T) {
	uc, mocks := NewTestLoginUsecase()
	mocks.userRepository.On("FindByUsername", mock.Anything).Return(nil, nil)
	// A bcrypt comparison must run even for a nonexistent account (against the
	// dummy hash) so the response time does not reveal that the user is missing.
	mocks.hashService.On("Compare", "password", mock.Anything).Return(false).Once()

	input := &usecase.LoginInput{
		Username: "username",
		Password: "password",
	}

	output, err := uc.Execute(input)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}

	mocks.AssertExpectations(t)
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	uc, mocks := NewTestLoginUsecase()
	mockUser := user.MustNew("User", "user01", "password")
	mockUser.SetID(1)
	mockUser.SetRoles(role.RoleAdmin)

	mocks.userRepository.On("FindByUsername", "user01").Return(mockUser, nil)
	mocks.hashService.On("Compare", "wrong_password", "password").Return(false)

	input := &usecase.LoginInput{
		Username: "user01",
		Password: "wrong_password",
	}

	output, err := uc.Execute(input)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}

	mocks.AssertExpectations(t)
}

func TestLoginUseCase_SaveRefreshTokenFailed(t *testing.T) {
	uc, mocks := NewTestLoginUsecase()
	mockUser := user.MustNew("User", "user01", "password")
	mockUser.SetID(1)
	mockUser.SetRoles(role.RoleAdmin)
	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "user01",
		Roles:           mockUser.GetRoles(),
		ExpirationTime:  int64(time.Hour.Seconds()),
	})
	refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
		UserID:         mockUser.GetID(),
		ExpirationTime: int64((time.Hour * 24 * 7).Seconds()),
		Token:          "refresh_token",
	})

	mocks.userRepository.On("FindByUsername", "user01").Return(mockUser, nil)
	mocks.hashService.On("Compare", "1234", "password").Return(true)
	mocks.tokenService.On("GenerateAccessToken", mockUser, int64(testAccessTTL)).Return(jwtToken, nil)
	mocks.tokenService.On("GenerateRefreshToken", mockUser, int64(testRefreshTTL)).Return(refreshToken, nil)
	mocks.refreshTokenRepository.On("Save", refreshToken).Return(errors.New("save failed"))

	input := &usecase.LoginInput{
		Username: "user01",
		Password: "1234",
	}

	output, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrInternalServerError) {
		t.Errorf("expected ErrInternalServerError, got %v", err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}

	mocks.AssertExpectations(t)
}
