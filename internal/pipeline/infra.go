package pipeline

import (
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/iac"
	_ "github.com/nguyentin05/cakd-platform/internal/iac/terraform"
)

// InfraStep is a pipeline step responsible for provisioning the application infrastructure
// (such as databases and repositories) using an Infrastructure as Code (IaC) provider like Terraform.
type InfraStep struct{}

// Name returns the step description.
func (s *InfraStep) Name() string { return "Creating Infrastructure (IaC)" }

// Run executes the IaC engine to provision cloud resources and caches the output variables
// in the shared pipeline context for downstream steps.
func (s *InfraStep) Run(ctx *Context) error {
	engine := iac.NewEngine(ctx.Cfg, ctx.OutDir)

	outputs, err := engine.Apply()
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	ctx.TfOutputs = outputs
	return nil
}
