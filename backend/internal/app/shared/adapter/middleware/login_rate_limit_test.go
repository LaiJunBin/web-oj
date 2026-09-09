package middleware

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLoginRateLimiter_TryConsumeAndRefund(t *testing.T) {
	l := NewLoginRateLimiter(3, 60) // burst 3, refill 1/sec
	now := time.Now()

	// The burst is consumable, then exhausted.
	for i := 0; i < 3; i++ {
		if !l.tryConsume("k", now) {
			t.Fatalf("consume %d within burst should succeed", i+1)
		}
	}
	if l.tryConsume("k", now) {
		t.Fatal("consume past the burst must fail")
	}

	// A refund (a request that did not fail auth) returns a token.
	l.refund("k", now)
	if !l.tryConsume("k", now) {
		t.Fatal("a refunded token must be consumable again")
	}
	if l.tryConsume("k", now) {
		t.Fatal("only one token was refunded")
	}

	// Refund never exceeds the burst ceiling.
	for i := 0; i < 10; i++ {
		l.refund("k", now)
	}
	got := 0
	for l.tryConsume("k", now) {
		got++
	}
	if got != 3 {
		t.Fatalf("refund must cap at burst; expected 3 consumable, got %d", got)
	}

	// One token refills after a second.
	if !l.tryConsume("k", now.Add(time.Second)) {
		t.Fatal("a token must refill after the interval")
	}

	// A different key is independent.
	if !l.tryConsume("other", now) {
		t.Fatal("a different key must have its own bucket")
	}
}

// A non-positive knob means "unlimited": tryConsume always succeeds, refund is a
// no-op.
func TestLoginRateLimiter_NonPositiveKnobIsUnlimited(t *testing.T) {
	now := time.Now()
	cases := []struct{ burst, perMinute int }{
		{0, 60}, {-1, 60}, {30, 0}, {30, -1}, {-5, -5},
	}
	for _, c := range cases {
		l := NewLoginRateLimiter(c.burst, c.perMinute)
		for i := 0; i < 1000; i++ {
			if !l.tryConsume("1.2.3.4", now) {
				t.Fatalf("burst=%d perMinute=%d must be unlimited, blocked at %d", c.burst, c.perMinute, i+1)
			}
		}
	}
}

// The whole point of reserve-and-refund: with a fixed clock (no refill), no more
// than burst reservations are ever handed out, no matter how many goroutines
// race for them. The old check-then-act model would let concurrent callers all
// pass on the same token and exceed burst — this pins that it cannot.
func TestLoginRateLimiter_ConcurrentConsumeBoundedByBurst(t *testing.T) {
	const burst = 50
	l := NewLoginRateLimiter(burst, 60)
	now := time.Now() // fixed: no refill for the duration of the test

	var admitted int64
	var wg sync.WaitGroup
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if l.tryConsume("same-key", now) {
					atomic.AddInt64(&admitted, 1)
				}
			}
		}()
	}
	wg.Wait()

	if admitted != burst {
		t.Fatalf("with a fixed clock exactly burst=%d may be admitted regardless of concurrency, got %d", burst, admitted)
	}
}

// tryConsume/refund run concurrently for every in-flight login. Hammer them from
// many goroutines — a real bucket (shared map + tokens under the mutex) and an
// unlimited one (whose lock-free l.unlimited read is the thing to prove is not a
// data race) — so `go test -race` can confirm it.
func TestLoginRateLimiter_ConcurrentNoRace(t *testing.T) {
	for _, l := range []*LoginRateLimiter{
		NewLoginRateLimiter(50, 60),
		NewLoginRateLimiter(0, 0),
	} {
		var wg sync.WaitGroup
		start := time.Now()
		for g := 0; g < 24; g++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				key := "ip-" + strconv.Itoa(id%5)
				for i := 0; i < 300; i++ {
					when := start.Add(time.Duration(i) * time.Millisecond)
					if l.tryConsume(key, when) {
						l.refund(key, when)
					}
				}
			}(g)
		}
		wg.Wait()
	}
}

// --- middleware integration ---

func newLoginTestEngine(perUser, perIP *LoginRateLimiter, exempt []*net.IPNet) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil) // ClientIP == RemoteAddr, deterministic
	r.Use(LoginRateLimit(perUser, perIP, exempt))
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		// Reads the body the middleware restored — a broken restore would fail
		// here and never return 200.
		_ = c.BindJSON(&body)
		if body.Password == "right" {
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
	})
	return r
}

func loginStatus(r *gin.Engine, ip, username, password string) int {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = ip + ":54321"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestLoginRateLimit_Middleware(t *testing.T) {
	t.Run("successful logins never consume quota", func(t *testing.T) {
		r := newLoginTestEngine(NewLoginRateLimiter(3, 60), NewLoginRateLimiter(100, 100), nil)
		for i := 0; i < 50; i++ {
			if code := loginStatus(r, "10.0.0.1", "alice", "right"); code != http.StatusOK {
				t.Fatalf("correct login %d should be 200, got %d — a room of successful logins must not lock itself out", i+1, code)
			}
		}
	})

	t.Run("failed logins are charged and blocked after the burst", func(t *testing.T) {
		r := newLoginTestEngine(NewLoginRateLimiter(3, 1), NewLoginRateLimiter(100, 1), nil)
		for i := 0; i < 3; i++ {
			if code := loginStatus(r, "10.0.0.2", "alice", "wrong"); code != http.StatusUnauthorized {
				t.Fatalf("failed login %d should be 401, got %d", i+1, code)
			}
		}
		if code := loginStatus(r, "10.0.0.2", "alice", "wrong"); code != http.StatusTooManyRequests {
			t.Fatalf("attempt past the burst should be 429, got %d", code)
		}
		// A different username from the same IP still has its own budget.
		if code := loginStatus(r, "10.0.0.2", "bob", "wrong"); code != http.StatusUnauthorized {
			t.Fatalf("a different username on the same IP should not be throttled, got %d", code)
		}
	})

	t.Run("per-IP backstop catches a username spray", func(t *testing.T) {
		// per-user effectively unlimited (huge burst); the backstop is what bites.
		r := newLoginTestEngine(NewLoginRateLimiter(1000, 1), NewLoginRateLimiter(3, 1), nil)
		for i := 0; i < 3; i++ {
			if code := loginStatus(r, "10.0.0.3", "user"+strconv.Itoa(i), "wrong"); code != http.StatusUnauthorized {
				t.Fatalf("spray attempt %d should be 401, got %d", i+1, code)
			}
		}
		if code := loginStatus(r, "10.0.0.3", "user-final", "wrong"); code != http.StatusTooManyRequests {
			t.Fatalf("spray past the per-IP backstop should be 429, got %d", code)
		}
	})

	t.Run("exempt CIDR skips the limiter", func(t *testing.T) {
		exempt, errs := ParseCIDRs([]string{"192.168.0.0/16"})
		if len(errs) != 0 {
			t.Fatalf("unexpected CIDR parse errors: %v", errs)
		}
		r := newLoginTestEngine(NewLoginRateLimiter(1, 1), NewLoginRateLimiter(1, 1), exempt)
		for i := 0; i < 20; i++ {
			if code := loginStatus(r, "192.168.1.5", "alice", "wrong"); code != http.StatusUnauthorized {
				t.Fatalf("exempt client must never be throttled, got %d at attempt %d", code, i+1)
			}
		}
	})
}
