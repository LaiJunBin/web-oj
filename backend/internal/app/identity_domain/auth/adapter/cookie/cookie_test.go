package cookie_test

import (
	"judge-backend/internal/app/identity_domain/auth/adapter/cookie"
	"judge-backend/internal/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func testConfig(cookieConfig *config.AuthCookieConfig) *config.Config {
	return &config.Config{Auth: &config.AuthConfig{Cookie: cookieConfig}}
}

// write runs fn against a throwaway gin context and returns the cookies it
// produced, parsed back out of the response header.
func write(t *testing.T, cfg *config.Config, fn func(k *cookie.Codec, c *gin.Context)) []*http.Cookie {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	fn(cookie.NewCodec(cfg), c)

	return (&http.Response{Header: w.Header()}).Cookies()
}

// read runs fn against a request carrying the given cookies.
func read(t *testing.T, cookies map[string]string, fn func(k *cookie.Codec, c *gin.Context) string) string {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	for name, value := range cookies {
		c.Request.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	return fn(cookie.NewCodec(testConfig(&config.AuthCookieConfig{Path: "/", SameSite: "Lax"})), c)
}

func only(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()

	if len(cookies) != 1 {
		t.Fatalf("expected exactly 1 cookie, but got %d", len(cookies))
	}
	return cookies[0]
}

func TestCodecSetAccessToken(t *testing.T) {
	cfg := testConfig(&config.AuthCookieConfig{
		Domain:   "example.com",
		Path:     "/app",
		Secure:   true,
		SameSite: "None",
	})

	got := only(t, write(t, cfg, func(k *cookie.Codec, c *gin.Context) {
		k.SetAccessToken(c, "value", 60)
	}))

	if got.Name != cookie.AccessToken || got.Value != "value" {
		t.Errorf("expected access_token=value, but got %s=%s", got.Name, got.Value)
	}
	if got.Path != "/app" {
		t.Errorf("expected path /app, but got %q", got.Path)
	}
	if got.Domain != "example.com" {
		t.Errorf("expected domain example.com, but got %q", got.Domain)
	}
	if !got.Secure {
		t.Error("expected Secure to be set")
	}
	if !got.HttpOnly {
		t.Error("expected HttpOnly to be set")
	}
	if got.MaxAge != 60 {
		t.Errorf("expected max age 60, but got %d", got.MaxAge)
	}

	// The attribute that had no call site at all before this: gin's
	// SetCookie cannot express it, so without SetSameSite every auth cookie
	// silently fell back to the browser default.
	if got.SameSite != http.SameSiteNoneMode {
		t.Errorf("expected SameSite=None, but got %v", got.SameSite)
	}
}

func TestCodecSetRefreshToken(t *testing.T) {
	cfg := testConfig(&config.AuthCookieConfig{Path: "/", SameSite: "Lax"})

	got := only(t, write(t, cfg, func(k *cookie.Codec, c *gin.Context) {
		k.SetRefreshToken(c, "value", 120)
	}))

	if got.Name != cookie.RefreshToken {
		t.Errorf("expected %s, but got %s", cookie.RefreshToken, got.Name)
	}
	if got.MaxAge != 120 {
		t.Errorf("expected max age 120, but got %d", got.MaxAge)
	}
}

func TestCodecSetAppliesEachSameSiteMode(t *testing.T) {
	for value, want := range map[string]http.SameSite{
		"Lax":    http.SameSiteLaxMode,
		"None":   http.SameSiteNoneMode,
		"Strict": http.SameSiteStrictMode,
	} {
		t.Run(value, func(t *testing.T) {
			cfg := testConfig(&config.AuthCookieConfig{Path: "/", SameSite: value})

			got := only(t, write(t, cfg, func(k *cookie.Codec, c *gin.Context) {
				k.SetAccessToken(c, "value", 60)
			}))

			if got.SameSite != want {
				t.Errorf("expected %v, but got %v", want, got.SameSite)
			}
		})
	}
}

// Every caller that clears clears both, so Clear does both and there is no
// way to leave one behind.
func TestCodecClearRemovesBothCookies(t *testing.T) {
	cfg := testConfig(&config.AuthCookieConfig{Path: "/", SameSite: "Lax"})

	got := write(t, cfg, func(k *cookie.Codec, c *gin.Context) {
		k.Clear(c)
	})

	if len(got) != 2 {
		t.Fatalf("expected 2 cookies, but got %d", len(got))
	}

	seen := map[string]*http.Cookie{}
	for _, ck := range got {
		seen[ck.Name] = ck
	}

	for _, name := range []string{cookie.AccessToken, cookie.RefreshToken} {
		ck, ok := seen[name]
		if !ok {
			t.Fatalf("expected %s to be cleared", name)
		}
		if ck.Value != "" {
			t.Errorf("expected %s to be emptied, but got %q", name, ck.Value)
		}
		if ck.MaxAge >= 0 {
			t.Errorf("expected %s to have a negative max age, but got %d", name, ck.MaxAge)
		}
	}
}

// A browser only drops a cookie when the expiring one matches the stored
// one's path and domain. Clearing on the defaults while login wrote a
// configured path would leave the session cookie in place — a logout that
// silently does not log out.
func TestCodecClearMatchesTheAttributesSetUsed(t *testing.T) {
	cfg := testConfig(&config.AuthCookieConfig{
		Domain:   "example.com",
		Path:     "/app",
		Secure:   true,
		SameSite: "Strict",
	})

	set := only(t, write(t, cfg, func(k *cookie.Codec, c *gin.Context) {
		k.SetRefreshToken(c, "value", 60)
	}))

	var cleared *http.Cookie
	for _, ck := range write(t, cfg, func(k *cookie.Codec, c *gin.Context) { k.Clear(c) }) {
		if ck.Name == cookie.RefreshToken {
			cleared = ck
		}
	}
	if cleared == nil {
		t.Fatal("expected the refresh token to be cleared")
	}

	if cleared.Path != set.Path || cleared.Domain != set.Domain {
		t.Errorf("expected clear to reuse path %q domain %q, but got path %q domain %q",
			set.Path, set.Domain, cleared.Path, cleared.Domain)
	}
	if cleared.Secure != set.Secure || cleared.SameSite != set.SameSite {
		t.Errorf("expected clear to reuse secure %v samesite %v, but got secure %v samesite %v",
			set.Secure, set.SameSite, cleared.Secure, cleared.SameSite)
	}
}

func TestCodecSetUsesTheConfiguredPathVerbatim(t *testing.T) {
	cfg := testConfig(&config.AuthCookieConfig{Path: "/", SameSite: "Lax"})

	got := only(t, write(t, cfg, func(k *cookie.Codec, c *gin.Context) {
		k.SetAccessToken(c, "value", 60)
	}))

	if got.Path != "/" {
		t.Errorf("expected path /, but got %q", got.Path)
	}
	if strings.Contains(got.String(), "Domain=") {
		t.Errorf("expected no Domain attribute for an empty domain, but got %q", got.String())
	}
}

// Attributes are read per call rather than captured when the codec is built,
// so a config changed at runtime takes effect without a restart — the same
// reason the usecases re-read their token lifetimes.
func TestCodecReadsAttributesPerCall(t *testing.T) {
	cookieConfig := &config.AuthCookieConfig{Path: "/", SameSite: "Lax"}
	cfg := testConfig(cookieConfig)

	gin.SetMode(gin.TestMode)
	k := cookie.NewCodec(cfg)

	cookieConfig.Path = "/changed"
	cookieConfig.Secure = true

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	k.SetAccessToken(c, "value", 60)

	got := only(t, (&http.Response{Header: w.Header()}).Cookies())
	if got.Path != "/changed" {
		t.Errorf("expected path /changed, but got %q", got.Path)
	}
	if !got.Secure {
		t.Error("expected Secure to follow the changed config")
	}
}

// Reading lives here too, so nothing outside this package needs to know what
// the auth cookies are called.
func TestCodecReadTokens(t *testing.T) {
	cookies := map[string]string{
		cookie.AccessToken:  "access-value",
		cookie.RefreshToken: "refresh-value",
		"unrelated":         "ignored",
	}

	if got := read(t, cookies, (*cookie.Codec).ReadAccessToken); got != "access-value" {
		t.Errorf("expected access-value, but got %q", got)
	}
	if got := read(t, cookies, (*cookie.Codec).ReadRefreshToken); got != "refresh-value" {
		t.Errorf("expected refresh-value, but got %q", got)
	}
}

func TestCodecReadTokensWhenAbsent(t *testing.T) {
	if got := read(t, map[string]string{}, (*cookie.Codec).ReadAccessToken); got != "" {
		t.Errorf("expected an empty string, but got %q", got)
	}
	if got := read(t, map[string]string{}, (*cookie.Codec).ReadRefreshToken); got != "" {
		t.Errorf("expected an empty string, but got %q", got)
	}
}
