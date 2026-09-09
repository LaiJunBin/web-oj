package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit caps how many bytes a request body may yield.
//
// It is the only place an upload's size is bounded. Multipart parsing spools
// file parts to temp files and FileHeader then reads each one whole into
// memory, so an unbounded request costs both disk and RAM before any handler
// gets a say. Capping the reader here bounds every route at once, and does
// it while the body is still being read rather than after it has landed.
//
// A non-positive limit disables the cap.
//
// overrides raises (or lowers) the cap for exact request paths. It exists for
// the bundle imports: an archive-mode image bundle carries docker save output
// and hundreds of megabytes is normal there, while every other upload endpoint
// stays at the general cap. Overriding here, where the general cap lives,
// keeps one place that answers "how big may this request be" — a second
// mechanism further in would be consulted only after this one had already
// closed the connection.
func BodyLimit(limit int64, overrides map[string]int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		effective := limit
		if o, ok := overrides[c.Request.URL.Path]; ok {
			effective = o
		}

		if effective > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, effective)
		}

		c.Next()
	}
}
