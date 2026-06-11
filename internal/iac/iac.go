package iac

import "github.com/nguyentin05/cakd-platform/internal/config"

// Engine defines the interface for provisioning and managing cloud infrastructure resources (IaC).
// Implementations of this interface manage resource lifecycles (e.g. using Terraform).
type Engine interface {
	// Apply provisions infrastructure resources and returns output variables (e.g. repo URLs).
	Apply() (outputs map[string]string, err error)
	// Destroy teardowns and cleans up all provisioned infrastructure resources.
	Destroy() error
}

// NewEngine is a factory function initialized by concrete IaC implementations to instantiate
// an [Engine] for a given configuration and working directory.
var NewEngine func(cfg *config.PlatformConfig, workDir string) Engine
