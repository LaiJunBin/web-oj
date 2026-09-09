package service

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
	"strconv"
)

type TokenService struct {
	jwtKey         string
	encryptService service.EncryptService
	randomService  service.RandomService
	logger         sharedtypes.Logger
}

func NewTokenService(
	config *config.Config,
	encryptService service.EncryptService,
	randomService service.RandomService,
	logger sharedtypes.Logger,
) service.TokenService {
	return &TokenService{
		jwtKey:         config.Server.JwtKey,
		encryptService: encryptService,
		randomService:  randomService,
		logger:         logger,
	}
}

func (ts *TokenService) GenerateAccessToken(user *user.User, expirationTime jwttoken.Second) (*jwttoken.JwtToken, error) {
	ts.logger.Debugf("Generating access token for user ID: %d, username: %s, expirationTime: %d", user.GetID(), user.GetUsername(), expirationTime)
	encryptedUserID, err := ts.encryptService.Encrypt(strconv.Itoa(user.GetID()))
	if err != nil {
		ts.logger.Errorf("Failed to encrypt user ID: %d, error: %v", user.GetID(), err)
		return nil, err
	}

	config := jwttoken.JwtTokenConfig{
		Secret:          ts.jwtKey,
		EncryptedUserID: encryptedUserID,
		Roles:           user.GetRoles(),
		ExpirationTime:  expirationTime,
	}
	jwtToken, err := jwttoken.NewJwtToken(config)
	if err != nil {
		ts.logger.Errorf("Failed to generate access token for user ID: %d, error: %v", user.GetID(), err)
		return nil, err
	}
	ts.logger.Debugf("Access token generated successfully for user ID: %d", user.GetID())
	return jwtToken, nil
}

func (ts *TokenService) ParseAccessToken(tokenString string) (*jwttoken.JwtToken, error) {
	ts.logger.Debugf("Parsing access token")
	token, err := jwttoken.NewJwtTokenWithRaw(ts.jwtKey, tokenString)
	if err != nil {
		ts.logger.Warnf("Failed to parse access token: %v", err)
		return nil, err
	}
	ts.logger.Debugf("Access token parsed successfully")
	return token, nil
}

func (ts *TokenService) GenerateRefreshToken(user *user.User, expirationTime refreshtoken.Second) (*refreshtoken.RefreshToken, error) {
	ts.logger.Debugf("Generating refresh token for user ID: %d, username: %s, expirationTime: %d", user.GetID(), user.GetUsername(), expirationTime)
	refreshToken, err := refreshtoken.NewRefreshToken(refreshtoken.RefreshTokenConfig{
		UserID:         user.GetID(),
		ExpirationTime: expirationTime,
		Token:          ts.randomService.RandString(64),
	})
	if err != nil {
		ts.logger.Errorf("Failed to generate refresh token for user ID: %d, error: %v", user.GetID(), err)
		return nil, err
	}
	ts.logger.Debugf("Refresh token generated successfully for user ID: %d", user.GetID())
	return refreshToken, nil
}
