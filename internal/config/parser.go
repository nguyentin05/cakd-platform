package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/nguyentin05/cakd-platform/internal/config/defaults"
	"github.com/nguyentin05/cakd-platform/internal/config/validate"
)

// Parse reads and processes the CAKD platform configuration YAML file.
// The parsing pipeline follows a strict, 3-phase architectural flow:
//  1. Structure Validation: Scans the raw YAML to ensure all required fields and enum
//     values are present, throwing errors before any mutation occurs.
//  2. Defaults Injection: Recursively applies implicit defaults (Convention over Configuration)
//     based on tags and internal registry values for missing fields.
//  3. Logic/Dependency Validation: Evaluates cross-field dependencies and business rules
//     (e.g., CI requires CD) now that the tree is fully hydrated with defaults.
func Parse(path string) (*PlatformConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var cfg PlatformConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if err := validate.Structure(&cfg); err != nil {
		return nil, err
	}

	defaults.Apply(&cfg)

	if err := validate.Logic(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
