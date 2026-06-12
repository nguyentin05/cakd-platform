package validate_test

import (
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config/validate"
	"github.com/nguyentin05/cakd-platform/internal/schema"
)

const (
	testAPIVersion = "platform.dev/v1alpha1"
	testKind       = "Project"
	testName       = "test"
	testOwner      = "test"
	testVC         = "github"
	testLang       = "java-spring-boot"
	testSvc1       = "svc1"
)

func TestStructure(t *testing.T) {
	tests := []struct {
		name        string
		cfg         schema.PlatformConfig
		expectError bool
		errContains string
	}{
		{
			name: "Valid - Fully omitted optional blocks",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Services: []schema.Service{
					{Name: testSvc1, Language: testLang},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid - Resources declared but empty",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Services: []schema.Service{
					{
						Name:      testSvc1,
						Language:  testLang,
						Resources: &schema.Resources{},
					},
				},
			},
			expectError: true,
			errContains: "block 'services[0].resources' is declared but empty",
		},
		{
			name: "Valid - Resources with only CPU",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Services: []schema.Service{
					{
						Name:      testSvc1,
						Language:  testLang,
						Resources: &schema.Resources{CPU: "500m"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid - Observability declared but empty",
			cfg: schema.PlatformConfig{
				APIVersion:    testAPIVersion,
				Kind:          testKind,
				Metadata:      schema.Metadata{Name: testName, Owner: testOwner},
				Providers:     schema.Providers{VersionControl: testVC},
				Observability: &schema.Observability{},
			},
			expectError: true,
			errContains: "block 'observability' is declared but empty",
		},
		{
			name: "Valid - Observability with Alerting enabled",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Observability: &schema.Observability{
					Alerting: true,
				},
			},
			expectError: false,
		},
		{
			name: "Invalid - AI Config declared but empty",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Observability: &schema.Observability{
					AI: &schema.AIConfig{},
				},
			},
			expectError: true,
			errContains: "block 'observability.ai' is declared but empty",
		},
		{
			name: "Valid - AI Config with Model",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Services: []schema.Service{
					{Name: testSvc1, Language: testLang},
				},
				Observability: &schema.Observability{
					AI: &schema.AIConfig{
						Model: "gemini-1.5-pro",
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid - Missing Required Field (Metadata.Name)",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Owner: testOwner}, // Missing Name
				Providers:  schema.Providers{VersionControl: testVC},
				Services: []schema.Service{
					{Name: testSvc1, Language: testLang},
				},
			},
			expectError: true,
			errContains: "field 'metadata.name' is required",
		},
		{
			name: "Invalid - Enum validation failed (Language)",
			cfg: schema.PlatformConfig{
				APIVersion: testAPIVersion,
				Kind:       testKind,
				Metadata:   schema.Metadata{Name: testName, Owner: testOwner},
				Providers:  schema.Providers{VersionControl: testVC},
				Services: []schema.Service{
					{Name: testSvc1, Language: "unknown-lang"},
				},
			},
			expectError: true,
			errContains: "unsupported language: \"unknown-lang\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Structure(&tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, but got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
