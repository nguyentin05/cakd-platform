package redact_test

import (
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/platform/redact"
)

func TestRedactMaskAndString(t *testing.T) {
	masked := redact.MaskSecret("dummy-secret-value-0000")
	if strings.Contains(masked, "secret-value") {
		t.Fatalf("expected secret middle to be masked, got %q", masked)
	}
	sanitized := redact.Sanitize("user=dummy-secret-value-0000", "dummy-secret-value-0000")
	if strings.Contains(sanitized, "dummy-secret-value-0000") {
		t.Fatalf("expected secret to be redacted from string, got %q", sanitized)
	}
}
