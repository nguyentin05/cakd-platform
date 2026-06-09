package template

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

//go:embed templates
var content embed.FS

type Engine struct {
	cfg *config.PlatformConfig
}

func New(cfg *config.PlatformConfig) *Engine {
	return &Engine{cfg: cfg}
}

func (e *Engine) Generate(outDir string) error {
	globalMappings := map[string]string{
		"templates/gitignore.tmpl": ".gitignore",
	}

	if e.cfg.Providers.CI != "" {
		globalMappings["templates/ci/ci.yml.tmpl"] = ".github/workflows/ci.yml"
	}

	if e.cfg.Providers.CD != "" {
		globalMappings["templates/argocd/application.yaml.tmpl"] = "deploy/application.yaml"
	}

	for tmplPath, outPath := range globalMappings {
		if err := e.renderTemplate(tmplPath, filepath.Join(outDir, outPath), e.cfg); err != nil {
			return fmt.Errorf("failed to render global %s: %w", tmplPath, err)
		}
	}

	for _, svc := range e.cfg.Services {
		svcDir := filepath.Join(outDir, svc.Name)
		svcMappings := map[string]string{
			"templates/dockerfile.tmpl": "Dockerfile",
		}

		if e.cfg.Providers.CD != "" {
			svcMappings["templates/helm/Chart.yaml.tmpl"] = "helm/Chart.yaml"
			svcMappings["templates/helm/values.yaml.tmpl"] = "helm/values.yaml"
			svcMappings["templates/helm/templates/deployment.yaml.tmpl"] = "helm/templates/deployment.yaml"
			svcMappings["templates/helm/templates/service.yaml.tmpl"] = "helm/templates/service.yaml"
		}

		if svc.Language == "java-spring-boot" {
			svcMappings["templates/spring-boot/src/main/resources/application.yml.tmpl"] = "src/main/resources/application.yml"

			usesDB := false
			for _, use := range svc.Uses {
				for _, b := range e.cfg.Backing {
					if b.Name == use && (b.Type == "postgresql" || b.Type == "mysql") {
						usesDB = true
						break
					}
				}
			}
			if usesDB {
				svcMappings["templates/spring-boot/src/test/resources/application.yml.tmpl"] = "src/test/resources/application.yml"
			}
		}

		type TemplateData struct {
			Config  *config.PlatformConfig
			Service config.Service
		}
		data := TemplateData{
			Config:  e.cfg,
			Service: svc,
		}

		for tmplPath, outPath := range svcMappings {
			if err := e.renderTemplate(tmplPath, filepath.Join(svcDir, outPath), data); err != nil {
				return fmt.Errorf("failed to render service template %s for %s: %w", tmplPath, svc.Name, err)
			}
		}
	}

	return nil
}

func (e *Engine) renderTemplate(tmplPath, outPath string, data interface{}) error {
	tmplContent, err := content.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("read embedded template %s: %w", tmplPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tmpl, err := template.New(filepath.Base(tmplPath)).Delims("[[", "]]").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
