package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/nguyentin05/cakd-platform/internal/config/defaults"
	"github.com/nguyentin05/cakd-platform/internal/config/validate"
	"github.com/nguyentin05/cakd-platform/internal/schema"
)

// Parse reads and processes the CAKD platform configuration YAML file.
// It executes a three-phase parsing pipeline:
//
// 1. Structure Validation: Ensures all required fields and enum values are present.
// 2. Defaults Injection: Traverses the configuration tree to hydrate missing fields with defaults.
// 3. Logic Validation: Evaluates cross-field dependencies and relationship constraints.
func Parse(path string) (*schema.PlatformConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var cfg schema.PlatformConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if err := validate.Structure(&cfg); err != nil {
		return nil, err
	}

	if err := defaults.Apply(&cfg); err != nil {
		return nil, fmt.Errorf("defaults injection failed: %w", err)
	}

	if err := validate.Logic(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
