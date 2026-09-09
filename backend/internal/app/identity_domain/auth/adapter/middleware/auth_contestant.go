package middleware

import (
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/gin-gonic/gin"
)

func AuthContestant(logger sharedtypes.Logger) api.MiddlewareFunc {
	return func(c *gin.Context) {
		logger.Debugf("AuthContestant middleware triggered")

		if _, exists := c.Get(api.PermissionContestantScopes); exists {
			logger.Debugf("PermissionContestantScopes found in context")
			user := c.MustGet(api.AuthorizationScopes).(*user.User)
			if user.HasPermission(role.PermissionContestant) {
				logger.Debugf("User %d has contestant permission", user.GetID())
				c.Next()
				return
			}

			logger.Warnf("User %d does not have contestant permission", user.GetID())
			c.JSON(403, api.SharedAPIForbiddenResponse{
				Message: "Forbidden",
			})
			c.Abort()
			return
		}

		logger.Debugf("PermissionContestantScopes not found, proceeding to next middleware")
		c.Next()
	}
}
