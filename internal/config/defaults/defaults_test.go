package defaults_test

import (
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/config/defaults"
)

func TestApply(t *testing.T) {
	cfg := &config.PlatformConfig{
		Services: []config.Service{
			{
				Name:     "web-api",
				Language: "java-spring-boot",
			},
		},
		Backing: []config.Backing{
			{
				Name: "main-db",
				Type: "postgresql",
			},
		},
		Providers: config.Providers{},
	}

	defaults.Apply(cfg)

	if cfg.Providers.VersionControl != "github" {
		t.Errorf("Expected VersionControl to be 'github', got %q", cfg.Providers.VersionControl)
	}

	svc := cfg.Services[0]
	if svc.LanguageVersion != "21" {
		t.Errorf("Expected Service LanguageVersion to be '21', got %q", svc.LanguageVersion)
	}
	if svc.Replicas != 1 {
		t.Errorf("Expected Replicas to be 1, got %d", svc.Replicas)
	}
	if svc.Resources == nil {
		t.Fatalf("Expected Resources to be initialized")
	}
	if svc.Resources.CPU != "500m" {
		t.Errorf("Expected CPU to be '500m', got %q", svc.Resources.CPU)
	}
	if svc.Resources.Memory != "512Mi" {
		t.Errorf("Expected Memory to be '512Mi', got %q", svc.Resources.Memory)
	}

	b := cfg.Backing[0]
	if b.Version != "16" {
		t.Errorf("Expected Backing Version to be '16', got %q", b.Version)
	}
	if b.Storage != "5Gi" {
		t.Errorf("Expected Storage to be '5Gi', got %q", b.Storage)
	}
}
