// Package cookie owns the auth cookies end to end: what they are called,
// what attributes they carry, and how they are read back.
//
// Keeping that in one place is not tidiness. Clearing a cookie only works
// when its attributes match the ones it was set with — a logout that expires
// access_token on path "/" will not touch a cookie the login wrote on path
// "/api", so the session survives a "successful" logout and no test notices
// because the response is still 200. Both the controller and the
// authorization middleware set and clear these cookies, so the attributes
// have to come from one place or they drift apart the first time the config
// changes.
//
// Reading lives here for the same reason: nothing outside this package
// should need to know that the tokens travel as "access_token" and
// "refresh_token".
package cookie

import (
	"judge-backend/internal/config"

	"github.com/gin-gonic/gin"
)

const (
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

// Codec reads and writes the auth cookies. Construct it through the DI
// container rather than inline, so every consumer shares one source of
// attributes.
type Codec struct {
	config *config.Config
}

func NewCodec(config *config.Config) *Codec {
	return &Codec{config: config}
}

// attributes are looked up per call rather than captured at construction, so
// a config changed at runtime takes effect without a restart — matching how
// the usecases re-read their token lifetimes.
func (k *Codec) attributes() *config.AuthCookieConfig {
	return k.config.Auth.Cookie
}

// set writes one auth cookie. httpOnly is always true: these tokens are
// never read by page scripts.
//
// SameSite has to be applied separately because gin's SetCookie predates the
// attribute and has no parameter for it; SetSameSite stashes the mode on the
// context for the next SetCookie call.
func (k *Codec) set(c *gin.Context, name string, value string, maxAge int) {
	attributes := k.attributes()
	c.SetSameSite(attributes.SameSiteMode())
	c.SetCookie(name, value, maxAge, attributes.Path, attributes.Domain, attributes.Secure, true)
}

func (k *Codec) SetAccessToken(c *gin.Context, value string, maxAge int) {
	k.set(c, AccessToken, value, maxAge)
}

func (k *Codec) SetRefreshToken(c *gin.Context, value string, maxAge int) {
	k.set(c, RefreshToken, value, maxAge)
}

// Clear expires both auth cookies, using the same attributes the setters
// would so the browser matches them against what it is actually holding.
//
// It deliberately takes no name: every caller clears both, and offering a
// per-cookie variant would only create the chance to leave one behind.
func (k *Codec) Clear(c *gin.Context) {
	k.set(c, AccessToken, "", -1)
	k.set(c, RefreshToken, "", -1)
}

func (k *Codec) read(c *gin.Context, name string) string {
	for _, ck := range c.Request.Cookies() {
		if ck.Name == name {
			return ck.Value
		}
	}
	return ""
}

// ReadAccessToken returns the presented access token, or "" if absent.
func (k *Codec) ReadAccessToken(c *gin.Context) string {
	return k.read(c, AccessToken)
}

// ReadRefreshToken returns the presented refresh token, or "" if absent.
func (k *Codec) ReadRefreshToken(c *gin.Context) string {
	return k.read(c, RefreshToken)
}
