package usecase

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

type RenewAccessTokenOutput struct {
	AccessToken Token
}

// RenewAccessTokenUseCase mints a fresh access token from a refresh token
// and leaves that refresh token exactly as it found it.
//
// It exists because RefreshTokenUseCase rotates: it mints a replacement
// refresh token and Replace()s the old one, which is a Redis RENAME that
// destroys the old key. That is correct for the explicit
// POST /auth/refresh-token endpoint, where the client makes one call and
// receives both cookies back.
//
// It is wrong for the authorization middleware. A contest page holds three
// SSE connections open at once, all carrying the same refresh cookie, so the
// moment the access token expires all three renew simultaneously: the first
// RENAME wins and the other two find their token gone, answer 401, and the
// middleware clears the contestant's cookies. The rotated refresh token
// could not have saved them either — it was issued on an SSE response behind
// a proxy and never reached the browser. Renewing without rotating makes
// concurrent renewals harmless by construction.
type RenewAccessTokenUseCase interface {
	Execute(refreshToken string) (*RenewAccessTokenOutput, error)
}

type renewAccessTokenUseCase struct {
	userRepository         repository.UserRepository
	refreshTokenRepository repository.RefreshTokenRepository
	tokenService           service.TokenService
	// config is held whole rather than having the lifetime copied out at
	// construction, so a value changed at runtime — or overridden per test
	// run — takes effect on the next renewal instead of at the next restart.
	config *config.Config
	logger sharedtypes.Logger
}

func NewRenewAccessTokenUseCase(
	userRepository repository.UserRepository,
	refreshTokenRepository repository.RefreshTokenRepository,
	tokenService service.TokenService,
	config *config.Config,
	logger sharedtypes.Logger,
) RenewAccessTokenUseCase {
	return &renewAccessTokenUseCase{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		config:                 config,
		logger:                 logger,
	}
}

func (uc *renewAccessTokenUseCase) Execute(refreshToken string) (*RenewAccessTokenOutput, error) {
	uc.logger.Debugf("Executing RenewAccessTokenUseCase")

	storedToken, err := uc.refreshTokenRepository.FindByToken(refreshToken)
	if err != nil {
		uc.logger.Errorf("Error finding refresh token: %v", err)
		return nil, ErrInternalServerError
	}

	if storedToken == nil {
		uc.logger.Warnf("Invalid refresh token")
		return nil, ErrInvalidToken
	}

	user, err := uc.userRepository.FindByID(storedToken.UserID().Value())
	if err != nil {
		uc.logger.Errorf("Error finding user by ID: %v", err)
		return nil, ErrInternalServerError
	}

	if user == nil {
		uc.logger.Warnf("Invalid credentials for user ID: %d", storedToken.UserID().Value())
		return nil, ErrInvalidCredentials
	}

	// The access lifetime comes from config, not from whatever the refresh
	// token has left. The refresh token is deliberately untouched from here
	// on: no GenerateRefreshToken, no Replace, no Save.
	accessToken, err := uc.tokenService.GenerateAccessToken(user, int64(uc.config.Auth.AccessTokenTTLSeconds))
	if err != nil {
		uc.logger.Errorf("Error generating access token: %v", err)
		return nil, ErrInternalServerError
	}

	uc.logger.Debugf("Successfully renewed access token for user ID: %d", user.GetID())
	return &RenewAccessTokenOutput{
		AccessToken: Token{Value: accessToken.String(), MaxAge: accessToken.MaxAge()},
	}, nil
}
