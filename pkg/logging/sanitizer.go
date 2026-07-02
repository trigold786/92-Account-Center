package logging

import "strings"

var sensitiveFields = map[string]bool{
	"password":       true,
	"password_hash":  true,
	"access_token":   true,
	"refresh_token":  true,
	"secret":         true,
	"authorization":  true,
	"credit_card":    true,
	"card_number":    true,
	"ssn":            true,
	"id_card":        true,
}

func SanitizeValue(key, value string) string {
	lower := strings.ToLower(key)
	for sensitive := range sensitiveFields {
		if strings.Contains(lower, sensitive) {
			return "***REDACTED***"
		}
	}
	return value
}
