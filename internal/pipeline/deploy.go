package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/provider"
)

// DeployStep registers the application with the CD provider.
type DeployStep struct{}

func (s *DeployStep) Name() string { return "Registering CD application" }

func (s *DeployStep) ShouldRun(ctx *Context) bool {
	return ctx.Cfg.Providers.CD != ""
}

func (s *DeployStep) Run(ctx *Context) error {
	cdProvider, err := provider.GetCDProvider(ctx.Cfg.Providers.CD)
	if err != nil {
		return err
	}

	manifestPath := filepath.Join(ctx.OutDir, "deploy", "application.yaml")
	if err := cdProvider.Register(manifestPath); err != nil {
		fmt.Fprintf(os.Stderr, "   %s registration failed: %v\n", ctx.Cfg.Providers.CD, err)
		fmt.Println("   Tip: Is your Minikube running? Do you have the CD tool installed?")
	} else {
		fmt.Printf("   %s application registered successfully\n", ctx.Cfg.Providers.CD)
	}

	return nil
}
