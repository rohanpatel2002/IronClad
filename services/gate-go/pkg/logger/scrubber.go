package logger

import (
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	tokenRegex = regexp.MustCompile(`(?i)(Bearer|token|secret|password|key)[:\s]+[a-zA-Z0-9._-]+`)
)

// ScrubPII removes sensitive information from a string.
func ScrubPII(input string) string {
	// Scrub emails
	output := emailRegex.ReplaceAllString(input, "[EMAIL_REDACTED]")
	
	// Scrub common tokens/secrets in logs
	output = tokenRegex.ReplaceAllStringFunc(output, func(m string) string {
		parts := strings.SplitN(m, ":", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(m, " ", 2)
		}
		if len(parts) > 1 {
			return parts[0] + " [REDACTED]"
		}
		return "[REDACTED]"
	})

	return output
}

// ScrubMap recursively scrubs PII from a map.
func ScrubMap(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for k, v := range input {
		switch val := v.(type) {
		case string:
			output[k] = ScrubPII(val)
		case map[string]interface{}:
			output[k] = ScrubMap(val)
		default:
			output[k] = v
		}
	}
	return output
}

