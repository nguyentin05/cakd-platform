package pipeline

import (
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

// Step is the smallest unit of work in the create pipeline.
type Step interface {
	Name() string
	Run(ctx *Context) error
}

// OptionalStep extends Step for steps that can be skipped based on config.
type OptionalStep interface {
	Step
	ShouldRun(ctx *Context) bool
}

// Execute runs the full create pipeline with the given config.
func Execute(cfg *config.PlatformConfig, force bool) error {
	ctx := NewContext(cfg, force)

	if err := ctx.Prepare(); err != nil {
		return err
	}

	steps := []Step{
		&ScaffoldStep{},
		&InfraStep{},
		&NotifyStep{},
		&VersionControlStep{},
		&DeployStep{},
	}

	if err := Run(ctx, steps); err != nil {
		return err
	}

	printSummary(ctx)
	return nil
}

// Run executes a list of steps sequentially.
func Run(ctx *Context, steps []Step) error {
	total := len(steps)
	for i, step := range steps {
		num := fmt.Sprintf("Step %d/%d", i+1, total)
		if opt, ok := step.(OptionalStep); ok && !opt.ShouldRun(ctx) {
			fmt.Printf("%s: Skipping %s (not configured)\n", num, step.Name())
			continue
		}
		fmt.Printf("%s: %s...\n", num, step.Name())
		if err := step.Run(ctx); err != nil {
			return fmt.Errorf("%s failed: %w", step.Name(), err)
		}
	}
	return nil
}

func printSummary(ctx *Context) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("Project created successfully!")
	if ctx.TfOutputs != nil {
		fmt.Printf("Repository: %s\n", ctx.TfOutputs["repo_html_url"])
	}
	fmt.Printf("Local code: %s\n", ctx.OutDir)
	if ctx.Cfg.Providers.CD == "argocd" {
		fmt.Println("Deployment: Managed by ArgoCD (GitOps)")
	}
	fmt.Println("═══════════════════════════════════════════════")
}
