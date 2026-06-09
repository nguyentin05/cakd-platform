package validate_test

import (
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/config/validate"
)

const testPostgres = "postgres-db"

func TestLogic(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.PlatformConfig
		expectError bool
		errContains string
	}{
		{
			name: "Valid - Services match Backing",
			cfg: config.PlatformConfig{
				Backing: []config.Backing{
					{Name: testPostgres},
				},
				Services: []config.Service{
					{
						Name: "api-service",
						Uses: []string{testPostgres},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid - Service uses unknown Backing",
			cfg: config.PlatformConfig{
				Backing: []config.Backing{
					{Name: "postgres-db"},
				},
				Services: []config.Service{
					{
						Name: "api-service",
						Uses: []string{"unknown-db"},
					},
				},
			},
			expectError: true,
			errContains: "references unknown backing resource \"unknown-db\"",
		},
		{
			name: "Invalid - Zero services",
			cfg: config.PlatformConfig{
				Backing: []config.Backing{
					{Name: "postgres-db"},
				},
				Services: []config.Service{},
			},
			expectError: true,
			errContains: "at least one service must be defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Logic(&tt.cfg)
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
