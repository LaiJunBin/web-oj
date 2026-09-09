package usecase

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	AccessToken  Token
	RefreshToken Token
}

type LoginUseCase interface {
	Execute(input *LoginInput) (*LoginOutput, error)
}

type loginUseCase struct {
	userRepository         repository.UserRepository
	refreshTokenRepository repository.RefreshTokenRepository
	tokenService           service.TokenService
	hashService            service.HashService
	// config is held whole rather than having the lifetimes copied out at
	// construction, so a value changed at runtime — or overridden per test
	// run — takes effect on the next login instead of at the next restart.
	config *config.Config
	logger sharedtypes.Logger
}

// dummyBcryptHash is a valid bcrypt hash (of a throwaway string) compared
// against when the account does not exist, so a failed login runs the same
// bcrypt cost whether or not the username is real — closing the timing oracle.
// It is not a secret: its only job is to make the comparison take real time.
const dummyBcryptHash = "$2a$10$qnSvyOUyMRL4nMQhdQoleuCv9hUzPC.NPE5OwomRzOPjivXtZGQwe"

func NewLoginUseCase(
	userRepository repository.UserRepository,
	refreshTokenRepository repository.RefreshTokenRepository,
	tokenService service.TokenService,
	hashService service.HashService,
	config *config.Config,
	logger sharedtypes.Logger,
) LoginUseCase {
	return &loginUseCase{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		hashService:            hashService,
		config:                 config,
		logger:                 logger,
	}
}

func (uc *loginUseCase) Execute(input *LoginInput) (*LoginOutput, error) {
	uc.logger.Infof("Executing LoginUseCase for username: %s", input.Username)

	if input.Username == "" || input.Password == "" {
		uc.logger.Warnf("Invalid input: username or password is empty")
		return nil, ErrBadRequest
	}

	user, err := uc.userRepository.FindByUsername(input.Username)
	if err != nil {
		uc.logger.Errorf("Error finding user by username: %v", err)
		return nil, ErrInternalServerError
	}

	if user == nil {
		// Spend the same bcrypt time on a nonexistent account so its absence is
		// not observable through a faster response.
		uc.hashService.Compare(input.Password, dummyBcryptHash)
		uc.logger.Warnf("Invalid credentials for username: %s", input.Username)
		return nil, ErrInvalidCredentials
	}

	if !uc.hashService.Compare(input.Password, user.GetPassword()) {
		uc.logger.Warnf("Invalid credentials for username: %s", input.Username)
		return nil, ErrInvalidCredentials
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user, int64(uc.config.Auth.AccessTokenTTLSeconds))
	if err != nil {
		uc.logger.Errorf("Error generating access token: %v", err)
		return nil, ErrInternalServerError
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken(user, int64(uc.config.Auth.RefreshTokenTTLSeconds))
	if err != nil {
		uc.logger.Errorf("Error generating refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	if err := uc.refreshTokenRepository.Save(refreshToken); err != nil {
		uc.logger.Errorf("Error saving refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	uc.logger.Infof("Successfully logged in user: %s", input.Username)
	return &LoginOutput{
		AccessToken:  Token{Value: accessToken.String(), MaxAge: accessToken.MaxAge()},
		RefreshToken: Token{Value: refreshToken.String(), MaxAge: refreshToken.ExpirationTime().Value()},
	}, nil
}
