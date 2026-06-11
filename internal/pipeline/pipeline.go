package pipeline

import (
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

// Step represents the smallest execute unit of work within the project creation pipeline.
type Step interface {
	// Name returns the human-readable description of the step.
	Name() string
	// Run executes the core logic of the step using the shared pipeline context.
	Run(ctx *Context) error
}

// OptionalStep extends [Step] to support conditional execution.
// Steps implementing this interface can be skipped dynamically depending on the user's config.
type OptionalStep interface {
	Step
	// ShouldRun returns true if the step should execute, or false if it should be skipped.
	ShouldRun(ctx *Context) bool
}

// Execute runs the full creation pipeline sequentially using the provided configuration.
// It sets up the workspace context, registers all build/deploy steps, and triggers execution.
// It prints a success summary upon clean completion.
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

// Run executes a sequence of pipeline steps. It coordinates logging step transitions,
// evaluates optional steps, and halts execution immediately if any step returns an error.
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

// printSummary displays a stylized summary block on standard output detailing the newly
// created project's git repository location, local code path, and deployment mode.
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
