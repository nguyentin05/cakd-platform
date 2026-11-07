package redact

import "strings"

const redactedPlaceholder = "[REDACTED]"

// MaskSecret returns a masked preview of a secret credential for safe display.
func MaskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// Sanitize replaces all occurrences of non-empty secrets in input with [REDACTED].
func Sanitize(input string, secrets ...string) string {
	out := input
	for _, s := range secrets {
		if len(s) >= 4 {
			out = strings.ReplaceAll(out, s, redactedPlaceholder)
		}
	}
	return out
}
