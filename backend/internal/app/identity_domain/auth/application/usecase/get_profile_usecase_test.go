package usecase_test

import (
	"errors"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	"judge-backend/internal/app/identity_domain/auth/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"
	"time"
)

type GetProfileUseCaseMocks struct {
	userRepository *mocks.UserRepository
	tokenService   *mocks.TokenService
	encryptService *mocks.EncryptService
	logger         sharedtypes.Logger
}

func NewTestGetProfileUsecase() (usecase.GetProfileUseCase, *GetProfileUseCaseMocks) {
	userRepository := &mocks.UserRepository{}
	tokenService := &mocks.TokenService{}
	encryptService := &mocks.EncryptService{}

	mocks := &GetProfileUseCaseMocks{
		userRepository: userRepository,
		tokenService:   tokenService,
		encryptService: encryptService,
		logger:         logger.NewMockLogger(),
	}
	usecase := usecase.NewGetProfileUseCase(mocks.userRepository, mocks.tokenService, mocks.encryptService, mocks.logger)
	return usecase, mocks
}

func (mocks *GetProfileUseCaseMocks) AssertExpectations(t *testing.T) {
	mocks.userRepository.AssertExpectations(t)
	mocks.tokenService.AssertExpectations(t)
	mocks.encryptService.AssertExpectations(t)
}

func TestGetProfileUseCase_GetProfileSuccess(t *testing.T) {
	uc, mocks := NewTestGetProfileUsecase()
	mockUser := user.MustNew("Test", "test", "hashed_password")
	mockUser.SetID(1)
	mockUser.SetRoles(role.RoleAdmin)
	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "encrypted_user_id",
		Roles:           mockUser.GetRoles(),
		ExpirationTime:  int64(time.Hour),
	})
	mocks.tokenService.On("ParseAccessToken", "test").Return(jwtToken, nil)
	mocks.userRepository.On("FindByID", mockUser.GetID()).Return(mockUser, nil)
	mocks.encryptService.On("Decrypt", "encrypted_user_id").Return("1", nil)

	input := &usecase.GetProfileInput{
		AccessToken: "test",
	}

	output, err := uc.Execute(input)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if output.Name != mockUser.GetName() {
		t.Errorf("expected %s, got %s", mockUser.GetName(), output.Name)
	}

	if output.Username != mockUser.GetUsername() {
		t.Errorf("expected %s, got %s", mockUser.GetUsername(), output.Username)
	}

	if len(output.Roles) != 1 {
		t.Errorf("expected 1, got %d", len(output.Roles))
	}

	if output.Roles[0] != role.RoleAdmin.Name() {
		t.Errorf("expected %s, got %s", role.RoleAdmin.Name(), output.Roles[0])
	}

	mocks.AssertExpectations(t)
}

func TestGetProfileUseCase_GetProfileFail_TokenExpired(t *testing.T) {
	uc, mocks := NewTestGetProfileUsecase()
	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "encrypted_user_id",
		Roles:           []*role.Role{},
		ExpirationTime:  -int64(time.Hour),
	})
	mocks.tokenService.On("ParseAccessToken", "test").Return(jwtToken, nil)

	input := &usecase.GetProfileInput{
		AccessToken: "test",
	}

	output, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrExpiredToken) {
		t.Errorf("expected '%v' error, got %v", usecase.ErrExpiredToken, err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}

	mocks.AssertExpectations(t)
}

func TestGetProfileUseCase_GetProfileFail_UserNotFound(t *testing.T) {
	uc, mocks := NewTestGetProfileUsecase()
	jwtToken := jwttoken.MustNewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          "secret",
		EncryptedUserID: "encrypted_user_id",
		Roles:           []*role.Role{},
		ExpirationTime:  int64(time.Hour),
	})
	mocks.tokenService.On("ParseAccessToken", "test").Return(jwtToken, nil)
	mocks.encryptService.On("Decrypt", "encrypted_user_id").Return("1", nil)
	mocks.userRepository.On("FindByID", 1).Return(nil, nil)

	input := &usecase.GetProfileInput{
		AccessToken: "test",
	}

	output, err := uc.Execute(input)
	if err == nil || !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Errorf("expected '%v' error, got %v", usecase.ErrInvalidCredentials, err)
	}

	if output != nil {
		t.Errorf("expected nil, got %v", output)
	}

	mocks.AssertExpectations(t)
}
