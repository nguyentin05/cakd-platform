package iac

import "github.com/nguyentin05/cakd-platform/internal/config"

// Engine provisions infrastructure resources.
type Engine interface {
	Apply() (outputs map[string]string, err error)
	Destroy() error
}

// NewEngine is a factory function initialized by the concrete IaC implementation.
var NewEngine func(cfg *config.PlatformConfig, workDir string) Engine
