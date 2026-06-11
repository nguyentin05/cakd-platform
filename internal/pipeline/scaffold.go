package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/provider"
	"github.com/nguyentin05/cakd-platform/internal/scaffold"
)

// ScaffoldStep generates the project structure and applies templates.
type ScaffoldStep struct{}

func (s *ScaffoldStep) Name() string { return "Scaffolding project structure" }

func (s *ScaffoldStep) Run(ctx *Context) error {
	for _, svc := range ctx.Cfg.Services {
		appProvider, err := provider.GetAppProvider(svc.Language)
		if err != nil {
			return err
		}

		svcOutDir := filepath.Join(ctx.OutDir, "services", svc.Name)
		if err := os.MkdirAll(svcOutDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for service %s: %w", svc.Name, err)
		}

		if err := appProvider.Scaffold(ctx.Cfg, svc, svcOutDir); err != nil {
			return fmt.Errorf("scaffolding failed for %s: %w", svc.Name, err)
		}
		os.Remove(filepath.Join(svcOutDir, "src", "main", "resources", "application.properties"))
		fmt.Printf("   %s base project generated (official)\n", svc.Language)
	}

	templatesMsg := "Dockerfile"
	if ctx.Cfg.Providers.CD != "" {
		templatesMsg += ", Helm, ArgoCD"
	}
	if ctx.Cfg.Providers.CI != "" {
		templatesMsg += ", CI"
	}
	fmt.Printf("   Applying CAKD templates (%s)...\n", templatesMsg)

	tmplEngine := scaffold.New(ctx.Cfg)
	if err := tmplEngine.Generate(ctx.OutDir); err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}
	fmt.Printf("   Project ready at: %s\n", ctx.OutDir)

	return nil
}
