package create

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/argocd"
	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/git"
	"github.com/nguyentin05/cakd-platform/internal/initializr"
	"github.com/nguyentin05/cakd-platform/internal/template"
	"github.com/nguyentin05/cakd-platform/internal/terraform"
)

func Run(cfg *config.PlatformConfig) error {
	fmt.Printf("Starting create for project: %s\n", cfg.Metadata.Name)

	outDir := filepath.Join(".", "out", cfg.Metadata.Name)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create out directory: %w", err)
	}

	fmt.Println("Step 1/4: Generating project...")

	if cfg.Spec.Language == "java-spring-boot" {
		fmt.Println("   Downloading base project from start.spring.io...")
		if err := initializr.Generate(cfg, outDir); err != nil {
			return fmt.Errorf("spring initializr failed: %w", err)
		}
		os.Remove(filepath.Join(outDir, "src", "main", "resources", "application.properties"))
		fmt.Println("   Spring Boot base project generated (official)")
	}

	fmt.Println("   Applying CAKD templates (Dockerfile, Helm, CI, ArgoCD)...")
	tmplEngine := template.New(cfg)
	if err := tmplEngine.Generate(outDir); err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}
	fmt.Printf("   Project ready at: %s\n", outDir)

	fmt.Println("Step 2/4: Creating GitHub repository via Terraform...")
	tfBridge := terraform.New(cfg, outDir)
	tfOutputs, err := tfBridge.Apply()
	if err != nil {
		return fmt.Errorf("terraform failed: %w", err)
	}

	fmt.Println("Step 3/4: Pushing code to repository...")
	ghToken := os.Getenv("GITHUB_TOKEN")
	if err := git.InitAndPush(outDir, tfOutputs.RepoCloneURL, ghToken); err != nil {
		fmt.Println("   Git failed, rolling back Terraform resources...")
		if destroyErr := tfBridge.Destroy(); destroyErr != nil {
			fmt.Fprintf(os.Stderr, "   Rollback also failed: %v\n", destroyErr)
		} else {
			fmt.Println("   Rollback successful — GitHub repo removed")
		}
		return fmt.Errorf("git operations failed: %w", err)
	}

	fmt.Println("Step 4/4: Registering ArgoCD application...")
	manifestPath := filepath.Join(outDir, "deploy", "application.yaml")
	if err := argocd.Register(manifestPath); err != nil {
		fmt.Fprintf(os.Stderr, "   ArgoCD registration failed: %v\n", err)
		fmt.Println("   Tip: Is your Minikube running? Do you have ArgoCD installed?")
	} else {
		fmt.Println("   ArgoCD application registered successfully")
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("Project created successfully!")
	fmt.Printf("Repository: %s\n", tfOutputs.RepoHTMLURL)
	fmt.Printf("Local code: %s\n", outDir)
	fmt.Println("Deployment: Managed by ArgoCD (GitOps)")
	fmt.Println("═══════════════════════════════════════════════")

	return nil
}
