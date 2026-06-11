package pipeline

import (
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/iac/terraform"
)

// InfraStep provisions infrastructure using Terraform.
type InfraStep struct{}

func (s *InfraStep) Name() string { return "Creating Infrastructure (IaC)" }

func (s *InfraStep) Run(ctx *Context) error {
	engine := terraform.New(ctx.Cfg, ctx.OutDir)

	outputs, err := engine.Apply()
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	ctx.TfOutputs = outputs
	return nil
}
