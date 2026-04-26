package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware injects an X-Request-ID header into the request and response
// if it's not already present. Useful for distributed tracing.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			reqID = hex.EncodeToString(b)
			c.Request.Header.Set("X-Request-ID", reqID)
		}

		// Set it on the response header too
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Next()
	}
}
