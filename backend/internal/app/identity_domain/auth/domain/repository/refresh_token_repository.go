package repository

import refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"

type RefreshTokenRepository interface {
	FindByToken(token string) (*refreshtoken.RefreshToken, error)
	Save(refreshToken *refreshtoken.RefreshToken) error
	Delete(refreshToken *refreshtoken.RefreshToken) error
	Replace(oldToken, newToken *refreshtoken.RefreshToken) error
	// DeleteAllByUserID revokes every session the user holds, for when the
	// credentials behind those sessions can no longer be trusted.
	DeleteAllByUserID(userID int) error
}
