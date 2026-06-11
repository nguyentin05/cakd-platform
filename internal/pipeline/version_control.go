package pipeline

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/iac/terraform"
	"github.com/nguyentin05/cakd-platform/internal/provider"
)

// VersionControlStep pushes the generated code to a remote repository.
type VersionControlStep struct{}

func (s *VersionControlStep) Name() string { return "Pushing code to repository" }

func (s *VersionControlStep) Run(ctx *Context) error {
	vcsProvider, err := provider.GetVersionControlProvider(ctx.Cfg.Providers.VersionControl)
	if err != nil {
		return err
	}

	ghToken := os.Getenv("GITHUB_TOKEN")

	if err := vcsProvider.InitAndPush(ctx.OutDir, ctx.TfOutputs["repo_clone_url"], ghToken); err != nil {
		fmt.Println("   Git failed, rolling back Terraform resources...")
		engine := terraform.New(ctx.Cfg, ctx.OutDir)
		if destroyErr := engine.Destroy(); destroyErr != nil {
			fmt.Fprintf(os.Stderr, "   Rollback also failed: %v\n", destroyErr)
		} else {
			fmt.Println("   Rollback successful — GitHub repo removed")
		}
		return fmt.Errorf("git operations failed: %w", err)
	}

	return nil
}
