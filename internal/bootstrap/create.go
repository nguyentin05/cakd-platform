package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/factory"
	"github.com/nguyentin05/cakd-platform/internal/provider/cd/argocd"
	"github.com/nguyentin05/cakd-platform/internal/template"
)

//nolint:gocyclo
func RunCreate(cfg *config.PlatformConfig, force bool) error {
	fmt.Printf("Starting create for project: %s\n", cfg.Metadata.Name)

	outDir := filepath.Join(".", "out", cfg.Metadata.Name)

	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		if !force {
			return fmt.Errorf("directory %s already exists. Use --force to overwrite", outDir)
		}
		if err := os.RemoveAll(outDir); err != nil {
			return fmt.Errorf("failed to clean out directory: %w", err)
		}
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create out directory: %w", err)
	}

	fac := factory.NewFactory(cfg)

	fmt.Println("Step 1/5: Scaffolding project structure...")
	for _, svc := range cfg.Services {
		appProvider, err := fac.GetAppFramework(svc)
		if err != nil {
			return fmt.Errorf("failed to get app framework provider for %s: %w", svc.Name, err)
		}

		svcOutDir := filepath.Join(outDir, svc.Name)
		if err := os.MkdirAll(svcOutDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for service %s: %w", svc.Name, err)
		}

		// Pass the whole cfg to Scaffold so the provider knows about Backing resources too.
		// Since Scaffold signature expects PlatformConfig, maybe we should also pass svc.Name so it knows which one.
		// Wait, Scaffold(cfg, outDir) is currently called. Let's just pass svcOutDir.
		if err := appProvider.Scaffold(cfg, svc, svcOutDir); err != nil {
			return fmt.Errorf("scaffolding project failed for %s: %w", svc.Name, err)
		}
		os.Remove(filepath.Join(svcOutDir, "src", "main", "resources", "application.properties"))
		fmt.Printf("   %s base project generated (official)\n", svc.Language)
	}

	fmt.Println("   Applying CAKD templates (Dockerfile, Helm, CI, ArgoCD)...")
	tmplEngine := template.New(cfg)
	if err := tmplEngine.Generate(outDir); err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}
	fmt.Printf("   Project ready at: %s\n", outDir)

	fmt.Println("Step 2/5: Creating Infrastructure (IaC)...")
	iacProvider, err := fac.GetIaC(outDir)
	if err != nil {
		return fmt.Errorf("failed to get IaC provider: %w", err)
	}

	tfOutputs, err := iacProvider.Apply()
	if err != nil {
		return fmt.Errorf("IaC apply failed: %w", err)
	}

	fmt.Println("Step 3/5: Provisioning Notification Channels...")
	notifier, err := fac.GetNotifier()
	if err != nil {
		fmt.Fprintf(os.Stderr, "   Warning: Failed to get notifier provider: %v\n", err)
	} else {
		webhookURL, err := notifier.ProvisionChannel(cfg.Metadata.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "   Warning: Failed to provision Discord channel: %v\n", err)
		} else {
			fmt.Printf("   Discord webhook created: %s\n", webhookURL)
			if err := config.SaveWebhook(cfg.Metadata.Name, webhookURL); err != nil {
				fmt.Fprintf(os.Stderr, "   Warning: Failed to save webhook to local config: %v\n", err)
			} else {
				fmt.Printf("   Webhook routing rule saved for namespace: %s\n", cfg.Metadata.Name)
			}
		}
	}

	fmt.Println("Step 4/5: Pushing code to repository...")
	vcsProvider, err := fac.GetVCS()
	if err != nil {
		return fmt.Errorf("failed to get VCS provider: %w", err)
	}

	ghToken := os.Getenv("GITHUB_TOKEN")
	if err := vcsProvider.InitAndPush(outDir, tfOutputs["repo_clone_url"], ghToken); err != nil {
		fmt.Println("   Git failed, rolling back Terraform resources...")
		if destroyErr := iacProvider.Destroy(); destroyErr != nil {
			fmt.Fprintf(os.Stderr, "   Rollback also failed: %v\n", destroyErr)
		} else {
			fmt.Println("   Rollback successful — GitHub repo removed")
		}
		return fmt.Errorf("git operations failed: %w", err)
	}

	fmt.Println("Step 5/5: Registering ArgoCD application...")
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
	fmt.Printf("Repository: %s\n", tfOutputs["repo_html_url"])
	fmt.Printf("Local code: %s\n", outDir)
	fmt.Println("Deployment: Managed by ArgoCD (GitOps)")
	fmt.Println("═══════════════════════════════════════════════")

	return nil
}
