package middleware_test

import (
	"bytes"
	"io"
	"judge-backend/internal/app/shared/adapter/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// readBodyUnder runs a request of bodySize bytes through BodyLimit(limit) and
// reports what the handler saw when it read the body. Reading is what the
// limit acts on — MaxBytesReader does not reject up front, it fails the read
// once the handler goes past the cap — so the handler's read error is the
// observable behaviour worth asserting.
func readBodyUnder(t *testing.T, limit int64, bodySize int) (read int, readErr error) {
	return readBodyUnderWithOverrides(t, limit, nil, "/", bodySize)
}

// readBodyUnderWithOverrides posts to path so the per-path override table can
// be exercised: the bundle imports carry docker save output and get a larger
// cap, and every other route must NOT inherit it.
func readBodyUnderWithOverrides(t *testing.T, limit int64, overrides map[string]int64, path string, bodySize int) (read int, readErr error) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.BodyLimit(limit, overrides))
	r.POST(path, func(c *gin.Context) {
		var body []byte
		body, readErr = io.ReadAll(c.Request.Body)
		read = len(body)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(make([]byte, bodySize)))
	r.ServeHTTP(httptest.NewRecorder(), req)

	return read, readErr
}

func TestBodyLimit(t *testing.T) {
	t.Run("rejects a body over the limit", func(t *testing.T) {
		read, err := readBodyUnder(t, 1024, 4096)

		if err == nil {
			t.Fatal("expected reading an oversized body to fail, but got nil")
		}

		// The point of capping the reader rather than checking afterwards:
		// the oversized body is never fully buffered.
		if int64(read) > 1024 {
			t.Errorf("expected at most 1024 bytes to be read, but got %d", read)
		}
	})

	t.Run("accepts a body exactly at the limit", func(t *testing.T) {
		read, err := readBodyUnder(t, 1024, 1024)

		if err != nil {
			t.Fatalf("expected a body at the limit to be readable, got %v", err)
		}

		if read != 1024 {
			t.Errorf("expected 1024 bytes to be read, but got %d", read)
		}
	})

	t.Run("accepts a body under the limit", func(t *testing.T) {
		read, err := readBodyUnder(t, 1024, 512)

		if err != nil {
			t.Fatalf("expected a body under the limit to be readable, got %v", err)
		}

		if read != 512 {
			t.Errorf("expected 512 bytes to be read, but got %d", read)
		}
	})

	// The escape hatch, matching ServerConfig.MaxUploadBytes returning a
	// negative value when the limit is switched off.
	t.Run("applies no limit when the limit is not positive", func(t *testing.T) {
		for _, limit := range []int64{0, -1} {
			read, err := readBodyUnder(t, limit, 4096)

			if err != nil {
				t.Fatalf("expected no limit for %d, got %v", limit, err)
			}

			if read != 4096 {
				t.Errorf("expected 4096 bytes to be read for limit %d, but got %d", limit, read)
			}
		}
	})
}

// The override exists for the two bundle-import routes, whose archives dwarf
// every other upload. The half that hurts if it silently breaks is the second
// one: an override that leaked to other paths would raise the cap for every
// endpoint at once, which is exactly what having a separate key was meant to
// avoid.
func TestBodyLimitPerPathOverride(t *testing.T) {
	overrides := map[string]int64{"/api/v1/bundles/import": 8192}

	t.Run("the overridden path takes the larger cap", func(t *testing.T) {
		read, err := readBodyUnderWithOverrides(t, 1024, overrides, "/api/v1/bundles/import", 4096)
		if err != nil {
			t.Fatalf("expected the override to admit 4096 bytes, but the read failed: %v", err)
		}
		if read != 4096 {
			t.Errorf("expected the whole body, but read %d", read)
		}
	})

	t.Run("every other path keeps the general cap", func(t *testing.T) {
		if _, err := readBodyUnderWithOverrides(t, 1024, overrides, "/api/v1/assets", 4096); err == nil {
			t.Error("expected the general cap to reject 4096 bytes, but the read succeeded")
		}
	})

	t.Run("an override below the general cap also applies", func(t *testing.T) {
		tight := map[string]int64{"/small": 512}
		if _, err := readBodyUnderWithOverrides(t, 1024, tight, "/small", 1000); err == nil {
			t.Error("expected the tighter override to reject 1000 bytes, but the read succeeded")
		}
	})
}
