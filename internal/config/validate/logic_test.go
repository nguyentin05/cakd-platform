package validate_test

import (
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config/validate"
	"github.com/nguyentin05/cakd-platform/internal/schema"
)

const testPostgres = "postgres-db"

func TestLogic(t *testing.T) {
	tests := []struct {
		name        string
		cfg         schema.PlatformConfig
		expectError bool
		errContains string
	}{
		{
			name: "Valid - Services match Backing",
			cfg: schema.PlatformConfig{
				Backing: []schema.Backing{
					{Name: testPostgres},
				},
				Services: []schema.Service{
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
			cfg: schema.PlatformConfig{
				Backing: []schema.Backing{
					{Name: "postgres-db"},
				},
				Services: []schema.Service{
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
			cfg: schema.PlatformConfig{
				Backing: []schema.Backing{
					{Name: "postgres-db"},
				},
				Services: []schema.Service{},
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
