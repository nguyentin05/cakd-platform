package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

// Context holds all shared state and runtime properties between pipeline steps.
// It acts as the single source of truth for step configurations and execution outputs.
type Context struct {
	// Cfg represents the parsed configuration of the platform.
	Cfg *config.PlatformConfig
	// OutDir is the target directory where all scaffolding files will be written.
	OutDir string
	// Force indicates whether to overwrite the target directory if it already exists.
	Force bool

	// TfOutputs caches the key-value outputs produced by Terraform during the InfraStep.
	// These values are typically consumed by downstream VCS and Deployment steps.
	TfOutputs map[string]string
}

// NewContext creates a new initialized pipeline Context with default output paths.
func NewContext(cfg *config.PlatformConfig, force bool) *Context {
	return &Context{
		Cfg:    cfg,
		OutDir: filepath.Join(".", "out", cfg.Metadata.Name),
		Force:  force,
	}
}

// Prepare sets up the output directory by validating if it exists, cleaning it
// if the force flag is active, and creating the directory path.
func (c *Context) Prepare() error {
	fmt.Printf("Starting create for project: %s\n", c.Cfg.Metadata.Name)

	if _, err := os.Stat(c.OutDir); !os.IsNotExist(err) {
		if !c.Force {
			return fmt.Errorf("directory %s already exists. Use --force to overwrite", c.OutDir)
		}
		if err := os.RemoveAll(c.OutDir); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	if err := os.MkdirAll(c.OutDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}
