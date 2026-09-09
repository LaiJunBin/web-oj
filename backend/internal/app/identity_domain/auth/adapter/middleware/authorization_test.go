package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/cookie"
	"judge-backend/internal/app/identity_domain/auth/adapter/middleware"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	"judge-backend/internal/app/identity_domain/auth/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// A token's "roles" claim is a snapshot taken at login. Authorising against it
// let a revoked role keep working for the rest of the access token's lifetime
// (default 3600s), which is exactly the window in which an admin is demoted
// mid-contest and expected to lose access at once. These tests pin the
// middleware to the roles the repository reports instead.

func newAdminClaimToken(t *testing.T, secret string) *jwttoken.JwtToken {
	t.Helper()

	token, err := jwttoken.NewJwtToken(jwttoken.JwtTokenConfig{
		Secret:          secret,
		EncryptedUserID: "encrypted-7",
		Roles:           []*role.Role{role.RoleAdmin},
		ExpirationTime:  3600,
	})
	if err != nil {
		t.Fatalf("build token: %v", err)
	}
	return token
}

// demotedUser is the same account the token was issued for, but the row now
// says it is only a contestant.
func demotedUser(t *testing.T) *user.User {
	t.Helper()

	u, err := user.New("Demoted Admin", "user07", "hashed")
	if err != nil {
		t.Fatalf("build user: %v", err)
	}
	if err := u.SetID(7); err != nil {
		t.Fatalf("set id: %v", err)
	}
	u.SetRoles(role.RoleContestant)
	return u
}

func runAuthorization(t *testing.T, scoped bool) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.NewTestConfig()
	log := logger.NewMockLogger()
	token := newAdminClaimToken(t, cfg.Server.JwtKey)

	tokenService := mocks.NewTokenService(t)
	tokenService.On("ParseAccessToken", mock.Anything).Return(token, nil).Maybe()

	encryptService := mocks.NewEncryptService(t)
	encryptService.On("Decrypt", "encrypted-7").Return("7", nil).Maybe()

	userRepository := mocks.NewUserRepository(t)
	userRepository.On("FindByID", 7).Return(demotedUser(t), nil).Maybe()

	renew := mocks.NewRenewAccessTokenUseCase(t)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.AddCookie(&http.Cookie{Name: "access_token", Value: "any"})
	if scoped {
		c.Set(api.AuthorizationScopes, []string{})
	}

	middleware.Authorization(
		tokenService, encryptService, userRepository, renew,
		cookie.NewCodec(cfg), log,
	)(c)

	return c, rec
}

func TestAuthorizationUsesRepositoryRolesNotTokenClaim(t *testing.T) {
	c, rec := runAuthorization(t, true)

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("authorization rejected a valid token: %d", rec.Code)
	}

	authed, ok := c.Get(api.AuthorizationScopes)
	if !ok {
		t.Fatal("authorization did not put the user on the context")
	}

	u, ok := authed.(*user.User)
	if !ok {
		t.Fatalf("context holds %T, want *user.User", authed)
	}

	// The token claim says admin; the row says contestant. The row wins.
	if u.HasPermission(role.PermissionAdmin) {
		t.Error("demoted user still has admin permission — roles were taken from the token claim")
	}
	if !u.HasPermission(role.PermissionContestant) {
		t.Error("user lost the contestant permission the repository reported")
	}
}

// Operations the spec leaves unauthenticated must stay untouched: the
// middleware runs on every route and self-gates on the scope key.
func TestAuthorizationSkipsOperationsWithoutScope(t *testing.T) {
	c, rec := runAuthorization(t, false)

	if _, exists := c.Get(api.AuthorizationScopes); exists {
		t.Error("middleware authorised an operation that declares no security")
	}
	if rec.Code == http.StatusUnauthorized {
		t.Error("middleware rejected a public operation")
	}
}

// clearingCookies returns the names of the auth cookies the response expires
// (Max-Age <= 0 with an empty value). A dead credential must be actively
// cleared; a browser holding a key signed by a rotated-away secret can
// otherwise never recover on its own.
func clearingCookies(rec *httptest.ResponseRecorder) map[string]bool {
	cleared := map[string]bool{}
	for _, ck := range (&http.Response{Header: rec.Header()}).Cookies() {
		if ck.Value == "" && ck.MaxAge <= 0 {
			cleared[ck.Name] = true
		}
	}
	return cleared
}

type authCase struct {
	presentCookies map[string]string
	parseErr       error
	renewOutput    *usecase.RenewAccessTokenOutput
	renewErr       error
	findByIDErr    error
}

func runAuthorizationCase(t *testing.T, tc authCase) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := config.NewTestConfig()
	log := logger.NewMockLogger()

	tokenService := mocks.NewTokenService(t)
	switch {
	case tc.renewOutput != nil:
		// Normal renewal: the presented access token is expired, renewal
		// succeeds, and the freshly minted token parses cleanly.
		tokenService.On("ParseAccessToken", mock.Anything).Return(nil, errors.New("token is expired")).Once()
		tokenService.On("ParseAccessToken", mock.Anything).Return(newAdminClaimToken(t, cfg.Server.JwtKey), nil).Maybe()
	case tc.parseErr != nil:
		tokenService.On("ParseAccessToken", mock.Anything).Return(nil, tc.parseErr).Maybe()
	default:
		tokenService.On("ParseAccessToken", mock.Anything).Return(newAdminClaimToken(t, cfg.Server.JwtKey), nil).Maybe()
	}

	encryptService := mocks.NewEncryptService(t)
	encryptService.On("Decrypt", "encrypted-7").Return("7", nil).Maybe()

	userRepository := mocks.NewUserRepository(t)
	if tc.findByIDErr != nil {
		userRepository.On("FindByID", 7).Return(nil, tc.findByIDErr).Maybe()
	} else {
		userRepository.On("FindByID", 7).Return(demotedUser(t), nil).Maybe()
	}

	renew := mocks.NewRenewAccessTokenUseCase(t)
	renew.On("Execute", mock.Anything).Return(tc.renewOutput, tc.renewErr).Maybe()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	for name, value := range tc.presentCookies {
		c.Request.AddCookie(&http.Cookie{Name: name, Value: value})
	}
	c.Set(api.AuthorizationScopes, []string{})

	middleware.Authorization(
		tokenService, encryptService, userRepository, renew,
		cookie.NewCodec(cfg), log,
	)(c)

	return rec
}

func signatureInvalid() error { return errors.New("token signature is invalid") }
func assertDBDown() error     { return errors.New("dial tcp: connection refused") }

// A dead-but-present credential (signature no longer verifies after a key
// rotation) must come back 401 AND clear both auth cookies, or the browser is
// locked out until the cookie expires on its own.
func TestAuthorizationClearsCookiesOnInvalidSignature(t *testing.T) {
	rec := runAuthorizationCase(t, authCase{
		presentCookies: map[string]string{"access_token": "dead", "refresh_token": "also-dead"},
		parseErr:       signatureInvalid(),
		renewErr:       usecase.ErrInvalidToken,
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	cleared := clearingCookies(rec)
	if !cleared["access_token"] || !cleared["refresh_token"] {
		t.Errorf("both auth cookies must be cleared, got %v", cleared)
	}
}

// The renewal path ran and found the refresh token genuinely gone (not a
// backing-store error): still a dead credential, still cleared.
func TestAuthorizationClearsCookiesWhenRefreshInvalid(t *testing.T) {
	rec := runAuthorizationCase(t, authCase{
		presentCookies: map[string]string{"access_token": "dead", "refresh_token": "gone"},
		parseErr:       signatureInvalid(),
		renewErr:       usecase.ErrInvalidToken,
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if cleared := clearingCookies(rec); !cleared["access_token"] || !cleared["refresh_token"] {
		t.Errorf("both auth cookies must be cleared, got %v", cleared)
	}
}

// Infrastructure failure during renewal (Redis unreachable) must NOT clear:
// a blink of the backing store would otherwise log the whole venue out.
func TestAuthorizationDoesNotClearOnRenewInfrastructureError(t *testing.T) {
	rec := runAuthorizationCase(t, authCase{
		presentCookies: map[string]string{"access_token": "dead", "refresh_token": "present"},
		parseErr:       signatureInvalid(),
		renewErr:       usecase.ErrInternalServerError,
	})

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("infrastructure failure must not be reported as 401, got %d", rec.Code)
	}
	if rec.Code < 500 {
		t.Errorf("want a 5xx for an infrastructure failure, got %d", rec.Code)
	}
	if cleared := clearingCookies(rec); len(cleared) != 0 {
		t.Errorf("a backing-store failure must clear nothing, got %v", cleared)
	}
}

// Infrastructure failure while loading the user for an otherwise-valid access
// token (DB down): not a credential problem, must not clear.
func TestAuthorizationDoesNotClearOnUserLookupInfrastructureError(t *testing.T) {
	rec := runAuthorizationCase(t, authCase{
		presentCookies: map[string]string{"access_token": "valid", "refresh_token": "present"},
		findByIDErr:    assertDBDown(),
	})

	if rec.Code < 500 {
		t.Errorf("want a 5xx for a DB failure, got %d", rec.Code)
	}
	if cleared := clearingCookies(rec); len(cleared) != 0 {
		t.Errorf("a DB failure must clear nothing, got %v", cleared)
	}
}

// An anonymous request carries no auth cookies; it is still 401 but there is
// nothing to clear and no Set-Cookie should be attached.
func TestAuthorizationDoesNotClearForAnonymousRequest(t *testing.T) {
	rec := runAuthorizationCase(t, authCase{
		presentCookies: map[string]string{},
		parseErr:       signatureInvalid(),
		renewErr:       usecase.ErrInvalidToken,
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
	if cleared := clearingCookies(rec); len(cleared) != 0 {
		t.Errorf("an anonymous 401 must not attach cookie-clearing headers, got %v", cleared)
	}
}

// The normal renewal path (expired access token, refresh still good) must be
// untouched: it authorises, clears nothing, and issues a fresh access cookie.
func TestAuthorizationNormalRenewalStillWorks(t *testing.T) {
	fresh := newAdminClaimToken(t, config.NewTestConfig().Server.JwtKey)
	rec := runAuthorizationCase(t, authCase{
		presentCookies: map[string]string{"access_token": "expired", "refresh_token": "good"},
		parseErr:       nil,
		renewOutput:    &usecase.RenewAccessTokenOutput{AccessToken: usecase.Token{Value: fresh.String(), MaxAge: 3600}},
	})

	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("normal renewal must not 401")
	}
	if cleared := clearingCookies(rec); len(cleared) != 0 {
		t.Errorf("a successful authorisation must clear nothing, got %v", cleared)
	}
}
