package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

// Context holds all shared state between pipeline steps.
type Context struct {
	Cfg    *config.PlatformConfig
	OutDir string
	Force  bool

	// TfOutputs holds Terraform outputs written by InfraStep and read by VersionControlStep.
	TfOutputs map[string]string
}

// NewContext creates a new pipeline context.
func NewContext(cfg *config.PlatformConfig, force bool) *Context {
	return &Context{
		Cfg:    cfg,
		OutDir: filepath.Join(".", "out", cfg.Metadata.Name),
		Force:  force,
	}
}

// Prepare sets up the output directory.
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
