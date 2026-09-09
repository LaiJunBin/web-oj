package middleware

import (
	"judge-backend/api"
	sharedtypes "judge-backend/internal/app/shared/types"

	"judge-backend/internal/config"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	csrf "github.com/utrack/gin-csrf"
)

func Csrf(
	config *config.Config,
	store redis.Store,
	logger sharedtypes.Logger,
) api.MiddlewareFunc {
	return func(c *gin.Context) {
		logger.Debugf("Csrf middleware triggered")

		sessions.Sessions("csrf", store)(c)
		csrf.Middleware(csrf.Options{
			Secret: config.Server.CsrfKey,
			ErrorFunc: func(c *gin.Context) {
				logger.Warnf("CSRF token mismatch")
				c.JSON(400, api.SharedAPIBadRequestResponse{
					Message: "CSRF token mismatch",
				})
				c.Abort()
			},
			TokenGetter: func(c *gin.Context) string {
				token := c.Request.Header.Get("csrf-token")
				logger.Debugf("Retrieved CSRF token from request header")
				return token
			},
		})(c)
		c.Next()
	}
}
