package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

func Parse(path string) (*PlatformConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var cfg PlatformConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	applyDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *PlatformConfig) {
	if cfg.Spec.Features.Monitoring == nil {
		t := true
		cfg.Spec.Features.Monitoring = &t
	}
	if cfg.Spec.Features.Alerting == nil {
		t := true
		cfg.Spec.Features.Alerting = &t
	}
	if cfg.Spec.Version == "" {
		cfg.Spec.Version = "21"
	}
	if cfg.Spec.Dependencies.Database != nil {
		if cfg.Spec.Dependencies.Database.Version == "" {
			cfg.Spec.Dependencies.Database.Version = "16"
		}
		if cfg.Spec.Dependencies.Database.Storage == "" {
			cfg.Spec.Dependencies.Database.Storage = "5Gi"
		}
	}
}

func validate(cfg *PlatformConfig) error {
	if cfg.APIVersion != "platform.dev/v1alpha1" {
		return fmt.Errorf("unsupported apiVersion: %q", cfg.APIVersion)
	}
	if cfg.Kind != "Project" {
		return fmt.Errorf("unsupported kind: %q", cfg.Kind)
	}
	if cfg.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if cfg.Metadata.Owner == "" {
		return fmt.Errorf("metadata.owner is required")
	}
	if cfg.Spec.Language == "" {
		return fmt.Errorf("spec.language is required")
	}

	supportedLanguages := map[string]bool{
		"java-spring-boot": true,
	}
	if !supportedLanguages[cfg.Spec.Language] {
		return fmt.Errorf("unsupported language: %q (supported: java-spring-boot)", cfg.Spec.Language)
	}

	supportedJavaVersions := map[string]bool{
		"17": true,
		"21": true,
	}
	if !supportedJavaVersions[cfg.Spec.Version] {
		return fmt.Errorf("unsupported java version: %q (supported: 17, 21)", cfg.Spec.Version)
	}

	if db := cfg.Spec.Dependencies.Database; db != nil {
		supportedDBs := map[string]bool{
			"postgresql": true,
		}
		if !supportedDBs[db.Type] {
			return fmt.Errorf("unsupported database type: %q (supported: postgresql)", db.Type)
		}

		storageRegex := regexp.MustCompile(`^[1-9][0-9]*[KMGTP]i?$`)
		if !storageRegex.MatchString(db.Storage) {
			return fmt.Errorf("invalid storage format: %q (e.g. 5Gi, 10Gi)", db.Storage)
		}
	}

	return nil
}
