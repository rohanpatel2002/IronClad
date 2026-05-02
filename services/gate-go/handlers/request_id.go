package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware injects X-Request-ID and X-Correlation-ID headers into the request and response
// if they are not already present. Useful for distributed tracing across services.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = generateID()
			c.Request.Header.Set("X-Request-ID", reqID)
		}

		corrID := c.GetHeader("X-Correlation-ID")
		if corrID == "" {
			// If no correlation ID exists, this request starts a new correlation chain
			corrID = reqID
			c.Request.Header.Set("X-Correlation-ID", corrID)
		}

		// Set them on the response header too
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Writer.Header().Set("X-Correlation-ID", corrID)
		c.Next()
	}
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
