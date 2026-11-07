package redact_test

import (
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/platform/redact"
)

func TestRedactMaskAndString(t *testing.T) {
	masked := redact.MaskSecret("ghp_1234567890abcdef")
	if strings.Contains(masked, "1234567890") {
		t.Fatalf("expected secret middle to be masked, got %q", masked)
	}
	sanitized := redact.Sanitize("token=ghp_1234567890abcdef", "ghp_1234567890abcdef")
	if strings.Contains(sanitized, "ghp_1234567890abcdef") {
		t.Fatalf("expected secret to be redacted from string, got %q", sanitized)
	}
}
