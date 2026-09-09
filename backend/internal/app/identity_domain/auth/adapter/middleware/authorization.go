package middleware

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/cookie"
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	sharedtypes "judge-backend/internal/app/shared/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// unauthorizedClear rejects a request whose credential is DEFINITELY dead —
// malformed, a signature that no longer verifies (a key rotation is the field
// case), or a renewal that ran and found the refresh token genuinely gone —
// and expires the auth cookies on the way out. A browser holding a cookie
// signed by a rotated-away secret can never recover on its own: the front end
// bounces it off the login page for holding a cookie and hides the logout
// button for having no profile, so without an active clear the operator is
// locked out until the cookie expires. The clear only reaches a browser that
// actually presented a credential; anonymous callers get unauthorizedAnonymous.
//
// CSRF is deliberately left alone, matching Logout (both go through the same
// cookie.Codec.Clear, which owns exactly the two auth cookies): the CSRF token
// is not a dead credential and the browser refreshes it on the next load.
func unauthorizedClear(c *gin.Context, cookies *cookie.Codec) {
	cookies.Clear(c)
	unauthorizedAnonymous(c)
}

// unauthorizedAnonymous rejects with 401 but attaches no Set-Cookie: the
// caller presented nothing to clear, and a public client should not be handed
// cookie-clearing headers it never asked for.
func unauthorizedAnonymous(c *gin.Context) {
	c.JSON(401, api.SharedAPIUnauthorizedResponse{
		Message: "Unauthorized",
	})
	c.Abort()
}

// serviceUnavailable rejects when authorization could not be DECIDED because a
// backing store (Redis, the database) failed. It must never clear cookies: a
// blink of Redis would otherwise log the entire venue out at once, which is
// precisely the opposite of what a transient failure should cause. 503 also
// keeps the front end off its login/logout state machine, which treats a 401
// as "your session is gone" — a backing-store hiccup is not that.
func serviceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, api.SharedAPIUnauthorizedResponse{
		Details: []string{},
		Message: "authorization temporarily unavailable",
		Success: false,
	})
	c.Abort()
}

func Authorization(
	tokenService service.TokenService,
	encryptService service.EncryptService,
	userRepository repository.UserRepository,
	renewAccessTokenUsecase usecase.RenewAccessTokenUseCase,
	cookies *cookie.Codec,
	logger sharedtypes.Logger,
) api.MiddlewareFunc {
	return func(c *gin.Context) {
		logger.Debugf("Authorization middleware triggered")

		if _, exists := c.Get(api.AuthorizationScopes); !exists {
			c.Next()
			return
		}

		// parseJwtToken returns a non-nil error ONLY for a backing-store
		// failure (the user lookup hitting a dead database). A malformed or
		// expired token, a bad payload, or a user the row genuinely lacks are
		// all "definitely invalid" and come back (nil, nil, nil) — the caller
		// clears cookies for those but must not for the error.
		parseJwtToken := func(token string) (*jwttoken.JwtToken, *user.User, error) {
			logger.Debugf("Parsing JWT token")
			tokenPayload, err := tokenService.ParseAccessToken(token)
			if err != nil {
				logger.Warnf("Failed to parse access token: %v", err)
				return nil, nil, nil
			}

			if err := tokenPayload.Validate(); err != nil {
				logger.Warnf("Invalid token payload: %v", err)
				return nil, nil, nil
			}

			decryptedUserID, err := encryptService.Decrypt(tokenPayload.UserID())
			if err != nil {
				logger.Warnf("Failed to decrypt user ID: %v", err)
				return nil, nil, nil
			}

			userID, _ := strconv.Atoi(decryptedUserID)
			user, err := userRepository.FindByID(userID)
			if err != nil {
				// The row could not be READ — a database failure, not a
				// missing user. The repository already tells the two apart
				// (RecordNotFound comes back as (nil, nil)).
				logger.Errorf("Failed to load user %d: %v", userID, err)
				return nil, nil, err
			}
			if user == nil {
				logger.Warnf("User not found for ID: %d", userID)
				return nil, nil, nil
			}

			logger.Debugf("Successfully parsed token and found user: %d", userID)
			return tokenPayload, user, nil
		}

		// Renews the access token ONLY. Rotating the refresh token here is
		// what broke production: a contest page holds three SSE connections
		// open at once, all carrying the same refresh cookie, so an expired
		// access token made all three renew at the same instant. Rotation
		// Replace()s the old refresh token — a Redis RENAME — so the first
		// renewal won and the other two found their token gone, answered 401
		// and had the contestant's cookies cleared below. The replacement
		// could not have reached the browser either: it was issued on an SSE
		// response behind a proxy. The explicit POST /auth/refresh-token
		// endpoint still rotates, which is safe because the client makes one
		// call and gets both cookies back.
		// tryRenewAccessToken returns a non-nil error ONLY for a backing-store
		// failure (usecase.ErrInternalServerError). A missing refresh cookie or
		// a refresh token the store genuinely does not hold
		// (ErrInvalidToken / ErrInvalidCredentials) come back ("", nil) —
		// definitely invalid, cookies get cleared; the error path must not.
		tryRenewAccessToken := func() (string, error) {
			logger.Debugf("Attempting to renew access token")
			refreshToken := cookies.ReadRefreshToken(c)
			if refreshToken == "" {
				logger.Warnf("No refresh token found")
				return "", nil
			}

			output, err := renewAccessTokenUsecase.Execute(refreshToken)
			if err != nil {
				if errors.Is(err, usecase.ErrInternalServerError) {
					logger.Errorf("Failed to renew access token (backing store): %v", err)
					return "", err
				}
				logger.Warnf("Refresh token is no longer valid: %v", err)
				return "", nil
			}

			logger.Debugf("Access token successfully renewed")
			cookies.SetAccessToken(c, output.AccessToken.Value, int(output.AccessToken.MaxAge))

			return output.AccessToken.Value, nil
		}

		token := cookies.ReadAccessToken(c)
		// Deliberately not reassigned by renewal: the refresh token is
		// untouched, so the one presented is still the live one, and Logout
		// and RefreshToken read it back off the context.
		refreshToken := cookies.ReadRefreshToken(c)
		// Whether the caller presented any credential at all is decided from
		// the REQUEST cookies, before renewal writes anything to the response:
		// it selects clearing (a dead credential to expire) versus a bare 401
		// (an anonymous caller with nothing to clear).
		hadCredential := token != "" || refreshToken != ""

		tokenPayload, user, infraErr := parseJwtToken(token)
		if infraErr != nil {
			serviceUnavailable(c)
			return
		}
		if tokenPayload == nil {
			logger.Warnf("Access token invalid, attempting to renew")
			token, infraErr = tryRenewAccessToken()
			if infraErr != nil {
				serviceUnavailable(c)
				return
			}
			tokenPayload, user, infraErr = parseJwtToken(token)
			if infraErr != nil {
				serviceUnavailable(c)
				return
			}
			if tokenPayload == nil {
				logger.Warnf("Authorization failed, unauthorized access")
				if hadCredential {
					unauthorizedClear(c, cookies)
				} else {
					unauthorizedAnonymous(c)
				}
				return
			}
		}

		// Roles come from the row just loaded by parseJwtToken, NOT from the
		// token's "roles" claim. The claim is a snapshot taken at login, so
		// authorising against it meant a revoked role kept working until the
		// access token expired — up to auth.accessTokenTTLSeconds (default
		// 3600s) of standing privilege after an admin was demoted, with no way
		// to cut it short. The repository already resolves roles from the
		// users table, so reading them off the entity costs nothing extra: the
		// lookup happens on every authorised request either way.
		logger.Debugf("Authorization successful for user: %d, roles: %v", user.GetID(), user.GetRoleNames())

		c.Set("access_token", token)
		c.Set("refresh_token", refreshToken)
		c.Set(api.AuthorizationScopes, user)
		c.Next()
	}
}
