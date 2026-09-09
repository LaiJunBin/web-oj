package service

import (
	refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
)

type TokenService interface {
	GenerateAccessToken(user *user.User, expirationTime int64) (*jwttoken.JwtToken, error)
	ParseAccessToken(tokenString string) (*jwttoken.JwtToken, error)
	GenerateRefreshToken(user *user.User, expirationTime int64) (*refreshtoken.RefreshToken, error)
}
