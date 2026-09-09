package middleware

import (
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/gin-gonic/gin"
)

func WithUserID(logger sharedtypes.Logger) api.MiddlewareFunc {
	return func(c *gin.Context) {
		logger.Debugf("WithUserID middleware triggered")

		if _, exists := c.Get(api.UserIDScopes); exists {
			logger.Debugf("UserIDScopes found in context")
			user := c.MustGet(api.AuthorizationScopes).(*user.User)
			userID := int32(user.GetID())
			logger.Debugf("Setting UserIDScopes to %d", userID)
			c.Set(api.UserIDScopes, userID)
		}

		c.Next()
	}
}
