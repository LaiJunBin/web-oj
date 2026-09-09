package usecase

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"strconv"
)

type GetProfileInput struct {
	AccessToken string
}

type GetProfileOutput struct {
	Name     string
	Username string
	Roles    []string
}

type GetProfileUseCase interface {
	Execute(input *GetProfileInput) (*GetProfileOutput, error)
}

type getProfileUseCase struct {
	userRepository repository.UserRepository
	tokenService   service.TokenService
	encryptService service.EncryptService
	logger         sharedtypes.Logger
}

func NewGetProfileUseCase(
	userRepository repository.UserRepository,
	tokenService service.TokenService,
	encryptService service.EncryptService,
	logger sharedtypes.Logger,
) GetProfileUseCase {
	return &getProfileUseCase{
		userRepository: userRepository,
		tokenService:   tokenService,
		encryptService: encryptService,
		logger:         logger,
	}
}

func (uc *getProfileUseCase) Execute(input *GetProfileInput) (*GetProfileOutput, error) {
	uc.logger.Infof("Executing GetProfileUseCase")

	jwtToken, err := uc.tokenService.ParseAccessToken(input.AccessToken)
	if err != nil {
		uc.logger.Errorf("Error parsing access token: %v", err)
		return nil, ErrInvalidToken
	}

	if jwtToken.IsExpired() {
		uc.logger.Warnf("Access token is expired")
		return nil, ErrExpiredToken
	}

	userID, err := uc.encryptService.Decrypt(jwtToken.UserID())
	if err != nil {
		uc.logger.Errorf("Error decrypting user ID: %v", err)
		return nil, ErrInternalServerError
	}

	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		uc.logger.Errorf("Error converting user ID to integer: %v", err)
		return nil, ErrInvalidToken
	}

	user, err := uc.userRepository.FindByID(intUserID)
	if err != nil {
		uc.logger.Errorf("Error finding user by ID: %v", err)
		return nil, ErrInternalServerError
	}

	if user == nil {
		uc.logger.Warnf("Invalid credentials for user ID: %d", intUserID)
		return nil, ErrInvalidCredentials
	}

	uc.logger.Infof("Successfully retrieved profile for user ID: %d", intUserID)
	return &GetProfileOutput{
		Name:     user.GetName(),
		Username: user.GetUsername(),
		Roles:    user.GetRoleNames(),
	}, nil
}
