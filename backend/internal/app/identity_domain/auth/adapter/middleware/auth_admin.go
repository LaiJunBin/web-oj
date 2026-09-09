package middleware

import (
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/gin-gonic/gin"
)

func AuthAdmin(logger sharedtypes.Logger) api.MiddlewareFunc {
	return func(c *gin.Context) {
		logger.Debugf("AuthAdmin middleware triggered")

		if _, exists := c.Get(api.PermissionAdminScopes); exists {
			logger.Debugf("PermissionAdminScopes found in context")
			user := c.MustGet(api.AuthorizationScopes).(*user.User)
			if user.HasPermission(role.PermissionAdmin) {
				logger.Debugf("User %d has admin permission", user.GetID())
				c.Next()
				return
			}

			logger.Warnf("User %d does not have admin permission", user.GetID())
			c.JSON(403, api.SharedAPIForbiddenResponse{
				Message: "Forbidden",
			})
			c.Abort()
			return
		}

		logger.Debugf("PermissionAdminScopes not found, proceeding to next middleware")
		c.Next()
	}
}
