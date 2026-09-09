package middleware

import (
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/cookie"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/gin-gonic/gin"
)

func WithRefreshToken(cookies *cookie.Codec, logger sharedtypes.Logger) api.MiddlewareFunc {
	return func(c *gin.Context) {
		logger.Debugf("WithRefreshToken middleware triggered")

		if _, exists := c.Get(api.RefreshTokenScopes); exists {
			logger.Debugf("RefreshTokenScopes found in context")
			refreshToken := cookies.ReadRefreshToken(c)
			logger.Debugf("Setting refresh_token in context")
			c.Set("refresh_token", refreshToken)
		}

		c.Next()
	}
}
