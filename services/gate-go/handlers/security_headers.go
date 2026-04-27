package handlers

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds standard security-hardening headers to every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent browsers from sniffing the MIME type away from that declared by the server
		c.Header("X-Content-Type-Options", "nosniff")

		// Protect against clickjacking by preventing the page from being rendered in an iframe
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS filtering in browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security (HSTS) - enforce HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content-Security-Policy (CSP) - restrict where resources can be loaded from
		// For a REST API, we can be very restrictive.
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none';")

		// Referrer-Policy
		c.Header("Referrer-Policy", "no-referrer")

		c.Next()
	}
}
