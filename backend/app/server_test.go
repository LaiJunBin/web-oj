package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"judge-backend/api"
	"judge-backend/internal/app/shared/adapter/middleware"
	"judge-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func intPtr(v int) *int { return &v }

type recoveryTestLogger struct {
	errorMessage string
}

func (l *recoveryTestLogger) Debugf(string, ...interface{}) {}
func (l *recoveryTestLogger) Infof(string, ...interface{})  {}
func (l *recoveryTestLogger) Warnf(string, ...interface{})  {}
func (l *recoveryTestLogger) Fatalf(string, ...interface{}) {}
func (l *recoveryTestLogger) Panicf(string, ...interface{}) {}
func (l *recoveryTestLogger) Errorf(message string, args ...interface{}) {
	l.errorMessage = fmt.Sprintf(message, args...)
}

func TestRecoveryHandlerReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := &recoveryTestLogger{}
	router := gin.New()
	router.Use(gin.CustomRecoveryWithWriter(io.Discard, recoveryHandler(logger)))
	router.GET("/panic", func(*gin.Context) {
		panic("boom")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}

	var response api.SharedAPIInternalServerErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode recovery response: %v", err)
	}
	if response.Success {
		t.Fatal("expected success to be false")
	}
	if response.Message != api.InternalServerError {
		t.Fatalf("expected message %q, got %q", api.InternalServerError, response.Message)
	}
	if response.Details != nil {
		t.Fatalf("expected details to be nil, got %#v", response.Details)
	}
	if !strings.Contains(logger.errorMessage, "panic recovered: boom") {
		t.Fatalf("expected panic to be logged, got %q", logger.errorMessage)
	}
	if !strings.Contains(logger.errorMessage, "goroutine") {
		t.Fatalf("expected stack trace to be logged, got %q", logger.errorMessage)
	}
}

// A rate-limited 429 must carry CORS headers, or the browser turns it into an
// opaque network error and the client cannot show "too many attempts". This
// pins the middleware order registerMiddleware installs: CORS ahead of the
// LoginRateLimit that aborts. Reorder them and this fails.
func TestRegisterMiddleware_RateLimited429CarriesCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const origin = "http://localhost:5173"
	cfg := &config.Config{Server: &config.ServerConfig{
		AllowedOrigins:          []string{origin},
		MaxUploadSizeMB:         100,
		LoginRateLimitBurst:     intPtr(1),
		LoginRateLimitPerMinute: intPtr(60),
	}}

	r := gin.New()
	registerMiddleware(r, cfg, &recoveryTestLogger{}, middleware.NewLoginRateLimiter(1, 60))
	// Stub the login route to fail so the limiter charges and then rejects.
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
	})

	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
			strings.NewReader(`{"username":"a","password":"b"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", origin)
		req.RemoteAddr = "10.9.9.9:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	// burst 1: the first failed attempt is 401, the next is 429.
	if got := send(); got.Code != http.StatusUnauthorized {
		t.Fatalf("first attempt should be 401, got %d", got.Code)
	}
	rejected := send()
	if rejected.Code != http.StatusTooManyRequests {
		t.Fatalf("second attempt should be 429, got %d", rejected.Code)
	}
	if got := rejected.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("429 must carry Access-Control-Allow-Origin=%q (CORS must run before the limiter), got %q", origin, got)
	}
}
