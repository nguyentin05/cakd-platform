package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

func TestParse(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		yamlContent string
		setupFile   bool
		filePath    string
		expectError bool
		errContains string
	}{
		{
			name:        "Error - File Not Found",
			setupFile:   false,
			filePath:    filepath.Join(tempDir, "non_existent.yaml"),
			expectError: true,
			errContains: "cannot read file",
		},
		{
			name: "Error - Invalid YAML Syntax",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
	invalid_indentation: true
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "invalid_syntax.yaml"),
			expectError: true,
			errContains: "invalid YAML",
		},
		{
			name: "Error - Structure Validation Failed (Missing Required Metadata Name)",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  owner: test-owner
providers:
  versionControl: github
services:
  - name: api
    language: java-spring-boot
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "missing_required.yaml"),
			expectError: true,
			errContains: "field 'metadata.name' is required",
		},
		{
			name: "Error - Logic Validation Failed (Unknown Backing)",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: test-project
  owner: test-owner
providers:
  versionControl: github
services:
  - name: api
    language: java-spring-boot
    uses:
      - unknown-db
backing:
  - name: known-db
    type: postgresql
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "logic_failed.yaml"),
			expectError: true,
			errContains: "references unknown backing resource \"unknown-db\"",
		},
		{
			name: "Error - Empty Option Block (Parent without child)",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: test-project
  owner: test-owner
providers:
  versionControl: github
services:
  - name: api
    language: java-spring-boot
observability: {}
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "empty_block.yaml"),
			expectError: true,
			errContains: "block 'observability' is declared but empty",
		},
		{
			name: "Error - Enum Validation Failed (Unallowed Value)",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: test-project
  owner: test-owner
providers:
  versionControl: github
services:
  - name: api
    language: rust
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "enum_failed.yaml"),
			expectError: true,
			errContains: "unsupported language: \"rust\"",
		},

		{
			name: "Error - Dependency Rule Failed (Alerting without Notification)",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: test-project
  owner: test-owner
providers:
  versionControl: github
observability:
  alerting: true
services:
  - name: api
    language: java-spring-boot
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "dep_failed.yaml"),
			expectError: true,
			errContains: "providers.notification is required when observability.alerting is enabled",
		},
		{
			name: "Success - Valid Config with Defaults Applied",
			yamlContent: `
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: test-project
  owner: test-owner
providers:
  versionControl: github
services:
  - name: api
    language: java-spring-boot
    uses:
      - main-db
backing:
  - name: main-db
    type: postgresql
`,
			setupFile:   true,
			filePath:    filepath.Join(tempDir, "valid_config.yaml"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				err := os.WriteFile(tt.filePath, []byte(tt.yamlContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
			}

			cfg, err := config.Parse(tt.filePath)

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
				if cfg == nil {
					t.Errorf("Expected valid config object, but got nil")
				} else {
					if len(cfg.Services) > 0 && cfg.Services[0].LanguageVersion == "" {
						t.Errorf("Expected defaults to be applied (e.g. Service LanguageVersion), but it was empty")
					}
				}
			}
		})
	}
}
