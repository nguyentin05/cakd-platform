package create

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/template"
)

func Run(cfg *config.PlatformConfig) error {
	fmt.Printf("Starting create for project: %s\n", cfg.Metadata.Name)

	outDir := filepath.Join(".", "out", cfg.Metadata.Name)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create out directory: %w", err)
	}

	fmt.Println("Generating templates...")
	tmplEngine := template.New(cfg)
	if err := tmplEngine.Generate(outDir); err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}
	fmt.Printf("   Generated code at: %s\n", outDir)

	// Step 2: Terraform Bridge
	fmt.Println("⏳ Step 2/4: Provisioning infrastructure via Terraform...")
	// TODO: Call terraform bridge

	// Step 3: Git Operations
	fmt.Println("⏳ Step 3/4: Pushing code to repository...")
	// TODO: Call git operations

	// Step 4: ArgoCD
	fmt.Println("⏳ Step 4/4: Registering ArgoCD application...")
	// TODO: Call ArgoCD integration

	fmt.Println("✅ Create completed successfully!")
	return nil
}
