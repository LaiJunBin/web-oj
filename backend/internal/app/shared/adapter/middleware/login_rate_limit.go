package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRateLimiter is a per-key token bucket guarding login against password
// brute-forcing. It is dependency-free (no golang.org/x/time) and self-cleaning.
//
// Only FAILED attempts spend tokens (see the middleware), so a room of people
// logging in successfully at once never depletes a bucket — the reason a shared
// NAT used to lock its own users out.
type LoginRateLimiter struct {
	mu           sync.Mutex
	buckets      map[string]*loginBucket
	refillPerSec float64
	burst        float64
	lastCleanup  time.Time
	// unlimited makes the limiter a pass-through, set when a knob was
	// non-positive so a misconfiguration can never lock users out.
	unlimited bool
}

type loginBucket struct {
	tokens float64
	last   time.Time
}

// NewLoginRateLimiter allows up to burst failed attempts, then one more every
// perMinute/60 seconds, per key.
func NewLoginRateLimiter(burst int, perMinute int) *LoginRateLimiter {
	l := &LoginRateLimiter{buckets: map[string]*loginBucket{}}
	if burst <= 0 || perMinute <= 0 {
		// A non-positive knob means "no limit" (the off/unlimited spelling the
		// config uses). Make the limiter a pass-through rather than build a
		// bucket that could lock users out — a non-positive burst would 429
		// everyone immediately, a non-positive per-minute would never refill and
		// permanently block a key after its burst. Callers normally skip
		// constructing the limiter when disabled; this is the safety net.
		l.unlimited = true
		return l
	}
	l.burst = float64(burst)
	l.refillPerSec = float64(perMinute) / 60.0
	return l
}

// bucketFor returns key's bucket brought up to date (created full if new). The
// caller must hold l.mu.
func (l *LoginRateLimiter) bucketFor(key string, now time.Time) *loginBucket {
	b, ok := l.buckets[key]
	if !ok {
		b = &loginBucket{tokens: l.burst, last: now}
		l.buckets[key] = b
		return b
	}
	b.tokens += now.Sub(b.last).Seconds() * l.refillPerSec
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now
	return b
}

// tryConsume atomically refills key's bucket and, if it has a whole token,
// takes one and reports true. Check and take happen under a single lock, so N
// concurrent requests cannot all pass on the same refilled token — the reason
// the long-term rate stays equal to the refill rate no matter the concurrency.
// The token is a RESERVATION held for the request's duration: kept for a failed
// attempt, handed back by refund for anything else.
func (l *LoginRateLimiter) tryConsume(key string, now time.Time) bool {
	if l.unlimited {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cleanup(now)
	b := l.bucketFor(key, now)
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// refund returns a token reserved by tryConsume, for a request that turned out
// not to be a failed login. Because a token is always reserved before the
// request proceeds, a bucket never overshoots, so there is nothing to clamp at
// zero (the old model's flooring is what forgave concurrent overshoot); the cap
// at burst is all that is needed.
func (l *LoginRateLimiter) refund(key string, now time.Time) {
	if l.unlimited {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.bucketFor(key, now)
	b.tokens++
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
}

// ForgetUser drops every bucket belonging to username (across all client IPs),
// so a contestant whose earlier failures emptied their bucket can log in
// immediately after staff reset their password, instead of waiting out the
// refill. Keys are "ip\x00username"; a linear scan is fine — the map is already
// small (10-minute reap) and this runs only on a password change.
func (l *LoginRateLimiter) ForgetUser(username string) {
	if l.unlimited {
		return
	}
	suffix := "\x00" + username
	l.mu.Lock()
	defer l.mu.Unlock()
	for k := range l.buckets {
		if strings.HasSuffix(k, suffix) {
			delete(l.buckets, k)
		}
	}
}

// cleanup drops buckets that have refilled and gone quiet, so the map does not
// grow without bound under a spray of distinct keys (an attacker rotating
// usernames). The per-IP backstop bounds how fast that can happen; this reaps
// what it leaves behind. Lazy (called under the lock) rather than a background
// goroutine, so there is nothing to stop.
func (l *LoginRateLimiter) cleanup(now time.Time) {
	if now.Sub(l.lastCleanup) < 10*time.Minute {
		return
	}
	l.lastCleanup = now
	for k, b := range l.buckets {
		if now.Sub(b.last) > 10*time.Minute {
			delete(l.buckets, k)
		}
	}
}

// maxLoginUsernameScanBytes bounds how much of a login body is read to pull the
// username out — a login JSON is well under this, and reading no further keeps a
// giant body from being buffered here before the handler even runs.
const maxLoginUsernameScanBytes = 16 << 10

// LoginRateLimit throttles failed logins per (client IP, username) with a
// looser per-IP backstop, and charges only failures (401), so legitimate
// logins — even a whole room at once behind one NAT — cost nothing. Clients in
// an exempt CIDR skip the limiter entirely. perUser/perIP being pass-through
// (disabled) is handled inside blocked/penalize.
func LoginRateLimit(perUser, perIP *LoginRateLimiter, exempt []*net.IPNet) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost || !strings.HasSuffix(c.Request.URL.Path, "/auth/login") {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if ipInAny(ip, exempt) {
			c.Next()
			return
		}

		username := readLoginUsername(c)
		ipKey := ip
		userKey := ip + "\x00" + username

		now := time.Now()
		// Reserve a token up front, atomically, so N concurrent requests cannot
		// all pass on one refilled token — the long-term rate is then exactly the
		// refill rate regardless of concurrency. Take the per-IP backstop FIRST so
		// a spraying IP short-circuits before a per-username bucket is even created
		// (bounding how many buckets a username-rotating attacker can force in).
		if !perIP.tryConsume(ipKey, now) {
			rejectLogin(c)
			return
		}
		if !perUser.tryConsume(userKey, now) {
			perIP.refund(ipKey, now) // the request is rejected; do not charge the IP for it
			rejectLogin(c)
			return
		}

		// Hand the reservations back for anything that is not a failed
		// authentication, so only real failed guesses cost a token. Deferred so a
		// panic — which the recovery middleware turns into a 500, not a 401 —
		// refunds too.
		defer func() {
			if c.Writer.Status() != http.StatusUnauthorized {
				rnow := time.Now()
				perUser.refund(userKey, rnow)
				perIP.refund(ipKey, rnow)
			}
		}()

		c.Next()
	}
}

func rejectLogin(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"success": false,
		"message": "too many failed login attempts, please try again later",
	})
}

// ipInAny reports whether ip is inside any of nets.
func ipInAny(ip string, nets []*net.IPNet) bool {
	if len(nets) == 0 {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, n := range nets {
		if n != nil && n.Contains(parsed) {
			return true
		}
	}
	return false
}

// ParseCIDRs parses CIDR strings into networks, returning the ones that parsed
// and the errors for the ones that did not, so a caller can log and skip the
// bad entries rather than fail startup over a typo.
func ParseCIDRs(cidrs []string) ([]*net.IPNet, []error) {
	var nets []*net.IPNet
	var errs []error
	for _, cidr := range cidrs {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		nets = append(nets, n)
	}
	return nets, errs
}

// readLoginUsername pulls "username" out of the JSON body and restores the body
// so the handler still reads it whole. Only the first maxLoginUsernameScanBytes
// are buffered — enough for any real login, and a hard bound on what a hostile
// oversized body can make this buffer.
func readLoginUsername(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	head, err := io.ReadAll(io.LimitReader(c.Request.Body, maxLoginUsernameScanBytes))
	// Restore what was read, followed by whatever remains, so the handler sees
	// the original body regardless of parse success.
	c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(head), c.Request.Body))
	if err != nil {
		return ""
	}

	var payload struct {
		Username string `json:"username"`
	}
	_ = json.Unmarshal(head, &payload)
	return payload.Username
}
