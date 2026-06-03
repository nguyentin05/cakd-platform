package create

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/git"
	"github.com/nguyentin05/cakd-platform/internal/template"
	"github.com/nguyentin05/cakd-platform/internal/terraform"
)

func Run(cfg *config.PlatformConfig) error {
	fmt.Printf("🚀 Starting create for project: %s\n", cfg.Metadata.Name)

	outDir := filepath.Join(".", "out", cfg.Metadata.Name)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create out directory: %w", err)
	}

	// Step 1: Generate Templates
	fmt.Println("⏳ Step 1/4: Generating templates...")
	tmplEngine := template.New(cfg)
	if err := tmplEngine.Generate(outDir); err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}
	fmt.Printf("   ✅ Generated code at: %s\n", outDir)

	// Step 2: Terraform — Create GitHub Repository
	fmt.Println("⏳ Step 2/4: Creating GitHub repository via Terraform...")
	tfBridge := terraform.New(cfg, outDir)
	tfOutputs, err := tfBridge.Apply()
	if err != nil {
		return fmt.Errorf("terraform failed: %w", err)
	}
	fmt.Printf("   ✅ Repository created: %s\n", tfOutputs.RepoHTMLURL)

	// Step 3: Git — Init, commit, push generated code
	fmt.Println("⏳ Step 3/4: Pushing code to repository...")
	ghToken := os.Getenv("GITHUB_TOKEN")
	if err := git.InitAndPush(outDir, tfOutputs.RepoCloneURL, ghToken); err != nil {
		// Rollback: destroy the GitHub repo so user can retry without "already exists" error
		fmt.Println("   ⚠️  Git failed, rolling back Terraform resources...")
		if destroyErr := tfBridge.Destroy(); destroyErr != nil {
			fmt.Fprintf(os.Stderr, "   ❌ Rollback also failed: %v\n", destroyErr)
		} else {
			fmt.Println("   ✅ Rollback successful — GitHub repo removed")
		}
		return fmt.Errorf("git operations failed: %w", err)
	}
	fmt.Printf("   ✅ Code pushed to: %s\n", tfOutputs.RepoHTMLURL)

	// Step 4: ArgoCD
	fmt.Println("⏳ Step 4/4: Registering ArgoCD application...")
	// TODO: Call ArgoCD integration (Bước 4)
	fmt.Println("   ⚠️  ArgoCD registration not yet implemented (Bước 4)")

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("✅ Project created successfully!")
	fmt.Printf("   📦 Repository : %s\n", tfOutputs.RepoHTMLURL)
	fmt.Printf("   📁 Local code : %s\n", outDir)
	fmt.Println("═══════════════════════════════════════════════")

	return nil
}
