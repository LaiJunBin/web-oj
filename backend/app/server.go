package app

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"judge-backend/api"
	"judge-backend/internal/app/shared/adapter/middleware"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Controller interface{}

type ServerImpl struct {
	api.StrictServerStub
}

// derefOrZero reads a *int config knob, treating an absent (nil) value as 0.
// Config.applyDefaults fills these in for real deployments; the nil guard keeps
// a hand-built config (mainly tests) from panicking.
func derefOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func recoveryHandler(logger sharedtypes.Logger) gin.RecoveryFunc {
	return func(c *gin.Context, err any) {
		logger.Errorf("panic recovered: %v\n%s", err, debug.Stack())
		c.AbortWithStatusJSON(http.StatusInternalServerError, api.SharedAPIInternalServerErrorResponse{
			Details: nil,
			Message: api.InternalServerError,
			Success: false,
		})
	}
}

// registerMiddleware installs the global middleware chain on r. It is extracted
// from SetupHandler so the ORDER — CORS ahead of the middlewares that can abort
// a request — can be tested without starting a server. That order is load
// bearing: see the CORS note below.
func registerMiddleware(r *gin.Engine, config *config.Config, logger sharedtypes.Logger, perUserLoginLimiter *middleware.LoginRateLimiter) {
	allowedOrigins := config.Server.AllowedOrigins

	// Decide whose X-Forwarded-For to believe. gin trusts ALL proxies by
	// default, which lets any client spoof its source IP — and with it slip past
	// the per-IP login limiter by sending a new fake address each request. An
	// empty allowlist (the default) trusts none, so ClientIP() is the real TCP
	// peer; behind a real proxy, set server.trustedProxies to that proxy.
	if err := r.SetTrustedProxies(config.Server.TrustedProxies); err != nil {
		logger.Errorf("invalid server.trustedProxies, refusing to trust any proxy: %v", err)
		_ = r.SetTrustedProxies(nil)
	}

	r.Use(gin.Logger())
	r.Use(gin.CustomRecoveryWithWriter(nil, recoveryHandler(logger)))

	// CORS is registered BEFORE the middlewares that can abort (BodyLimit's 413,
	// LoginRateLimit's 429) so every response — aborts included — carries the
	// access-control headers. Registered after them, an abort would answer before
	// CORS ran, and the browser would turn the header-less response into an opaque
	// network error: the client could not tell "rate limited" from "server
	// unreachable". Being early also lets a preflight OPTIONS short-circuit before
	// the heavier middlewares run.
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Cookie", "Csrf-Token", "Content-Type"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			logger.Debugf("CORS origin check: %s", origin)
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
		MaxAge: 12 * time.Hour,
	}))

	r.Use(metrics.Middleware())

	// Ahead of routing so it covers every upload endpoint, present and
	// future, rather than each handler having to remember.
	r.Use(middleware.BodyLimit(config.Server.MaxUploadBytes(), map[string]int64{
		// The two endpoints that legitimately receive an archive-mode image
		// bundle, which is docker save output and dwarfs every other upload.
		"/api/v1/bundles/import": config.Server.BundleImportMaxUploadBytes(),
		"/api/v1/images/import":  config.Server.BundleImportMaxUploadBytes(),
	}))

	// Per-IP brute-force guard on the login endpoint (configurable; keyed on the
	// real client IP thanks to SetTrustedProxies above). A non-positive burst OR
	// per-minute — the "off/unlimited" spelling the upload/pids/memory caps use —
	// disables it entirely, rather than the lock-everyone a non-positive burst
	// would otherwise produce. Only failed logins are charged, keyed per
	// (IP, username) with a fixed per-IP backstop — see the config docs.
	burst := derefOrZero(config.Server.LoginRateLimitBurst)
	perMinute := derefOrZero(config.Server.LoginRateLimitPerMinute)
	if burst > 0 && perMinute > 0 {
		// The looser per-IP backstop catches a single source spraying many
		// different usernames — which the per-(IP,username) buckets alone would
		// not, since each fresh username starts a fresh bucket. It is
		// configurable (a non-positive value makes it a pass-through), sized by
		// config to absorb a whole room's finger-slips without locking anyone out.
		ipBurst := derefOrZero(config.Server.LoginRateLimitIpBurst)
		ipPerMinute := derefOrZero(config.Server.LoginRateLimitIpPerMinute)
		exempt, errs := middleware.ParseCIDRs(config.Server.LoginRateLimitExemptCIDRs)
		for _, e := range errs {
			logger.Errorf("ignoring invalid server.loginRateLimitExemptCIDRs entry: %v", e)
		}
		// The per-user limiter is the SHARED instance (also injected into the
		// password-reset usecase, which clears a user's record via it), so both
		// act on the same buckets.
		r.Use(middleware.LoginRateLimit(
			perUserLoginLimiter,
			middleware.NewLoginRateLimiter(ipBurst, ipPerMinute),
			exempt,
		))
	}
}

func SetupHandler(s api.ServerInterface, config *config.Config, logger sharedtypes.Logger, m []api.MiddlewareFunc, perUserLoginLimiter *middleware.LoginRateLimiter) {
	port := strconv.Itoa(config.Server.Port)
	r := gin.New()
	registerMiddleware(r, config, logger, perUserLoginLimiter)
	api.RegisterHandlersWithOptions(r, s, api.GinServerOptions{
		BaseURL:     "/api/v1",
		Middlewares: m,
	})

	// Explicit server with a header-read deadline and idle reaping so a
	// slowloris client cannot hold connections open indefinitely on the header
	// phase, and dead keep-alives are cleaned up. Deliberately NO WriteTimeout
	// or ReadTimeout: the API serves long-lived SSE streams (a write deadline
	// would sever them) and accepts multi-hundred-MB bundle uploads whose body
	// read can outlast any reasonable ReadTimeout — body size is bounded by the
	// BodyLimit middleware instead.
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Errorf("http server stopped: %v", err)
	}
}

func NewServer(
	ctls []Controller,
) api.ServerInterface {
	server := &ServerImpl{}
	server.InitControllers(ctls)

	return api.NewStrictHandler(server, nil)
}
