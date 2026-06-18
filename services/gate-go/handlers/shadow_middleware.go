package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ShadowMiddleware asynchronously mirrors requests to a shadow environment.
func ShadowMiddleware() gin.HandlerFunc {
	shadowURL := os.Getenv("SHADOW_URL")
	if shadowURL == "" {
		return func(c *gin.Context) { c.Next() }
	}

	client := &http.Client{}

	return func(c *gin.Context) {
		// Clone request body for shadowing
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Asynchronously shadow the request
		go func(method, path string, body []byte, headers http.Header) {
			shadowReq, err := http.NewRequest(method, shadowURL+path, bytes.NewBuffer(body))
			if err != nil {
				return
			}

			// Copy headers
			for k, vv := range headers {
				for _, v := range vv {
					shadowReq.Header.Add(k, v)
				}
			}

			resp, err := client.Do(shadowReq)
			if err != nil {
				slog.Debug("Shadow request failed", "error", err)
				return
			}
			defer resp.Body.Close()
			slog.Debug("Shadow request sent", "status", resp.Status)
		}(c.Request.Method, c.Request.URL.Path, bodyBytes, c.Request.Header)

		c.Next()
	}
}

