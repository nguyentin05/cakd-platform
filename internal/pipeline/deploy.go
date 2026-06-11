package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/provider"
)

// DeployStep is an optional pipeline step responsible for registering the generated GitOps manifests
// with a Continuous Deployment (CD) provider like ArgoCD.
type DeployStep struct{}

// Name returns the step description.
func (s *DeployStep) Name() string { return "Registering CD application" }

// ShouldRun returns true if a CD provider is configured in platform.yaml.
func (s *DeployStep) ShouldRun(ctx *Context) bool {
	return ctx.Cfg.Providers.CD != ""
}

// Run registers the CD application using the generated deployment manifest file.
// If registration fails, it prints troubleshooting tips and continues without failing the pipeline.
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
