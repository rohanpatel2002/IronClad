package logger

import (
	"reflect"
	"testing"
)

func TestScrubPII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Scrub email",
			input:    "Contact me at user@example.com for more info.",
			expected: "Contact me at [EMAIL_REDACTED] for more info.",
		},
		{
			name:     "Scrub token with space",
			input:    "Authorization: Bearer 12345abcdef",
			expected: "Authorization: Bearer [REDACTED]",
		},
		{
			name:     "Scrub secret with colon",
			input:    "DB_PASSWORD:secretPassword123",
			expected: "DB_PASSWORD [REDACTED]", // based on current naive tokenRegex logic
		},
		{
			name:     "No PII",
			input:    "This is a normal log message",
			expected: "This is a normal log message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScrubPII(tt.input); got != tt.expected {
				t.Errorf("ScrubPII() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestScrubMap(t *testing.T) {
	input := map[string]interface{}{
		"user": "test@domain.com",
		"meta": map[string]interface{}{
			"auth": "Bearer secret_token",
			"id":   123,
		},
		"status": "active",
	}

	expected := map[string]interface{}{
		"user": "[EMAIL_REDACTED]",
		"meta": map[string]interface{}{
			"auth": "Bearer [REDACTED]",
			"id":   123,
		},
		"status": "active",
	}

	got := ScrubMap(input)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ScrubMap() = %v, want %v", got, expected)
	}
}
