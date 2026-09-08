package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohanpatel2002/ironclad/services/gate-go/pkg/threat"
)

var (
	// Basic regex for SQL injection patterns
	sqliRegex = regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|--|OR\s+1=1)`)
	// Basic regex for XSS patterns
	xssRegex = regexp.MustCompile(`(?i)(<script|alert\(|onerror=)`)
)

// WAFMiddleware provides lightweight protection against common web attacks.
func WAFMiddleware(intel *threat.IntelClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Global Threat Intel
		if intel.IsMalicious(c.ClientIP()) {
			RecordThreatBlockMetric()
			blockRequest(c, "request from known malicious IP")
			return
		}

		// Check URL parameters
		for _, values := range c.Request.URL.Query() {
			for _, v := range values {
				if sqliRegex.MatchString(v) || xssRegex.MatchString(v) {
					blockRequest(c, "injection attempt in query")
					return
				}
			}
		}

		// Check headers (custom inspection for User-Agent or others)
		ua := c.GetHeader("User-Agent")
		if strings.Contains(strings.ToLower(ua), "sqlmap") || strings.Contains(strings.ToLower(ua), "nikto") {
			blockRequest(c, "malicious tool detected")
			return
		}

		c.Next()
	}
}

func blockRequest(c *gin.Context, reason string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":   "security_violation",
		"message": "request blocked by IronClad WAF",
		"reason":  reason,
	})
}

// InspectInputString evaluates whether arbitrary input contains malicious SQLi or XSS patterns.
func InspectInputString(input string) (isMalicious bool, patternType string) {
	if sqliRegex.MatchString(input) {
		return true, "SQLi"
	}
	if xssRegex.MatchString(input) {
		return true, "XSS"
	}
	return false, ""
}

