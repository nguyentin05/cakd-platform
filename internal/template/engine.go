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
	mappings := map[string]string{
		"templates/gitignore.tmpl":                      ".gitignore",
		"templates/dockerfile.tmpl":                     "Dockerfile",
		"templates/ci/ci.yml.tmpl":                      ".github/workflows/ci.yml",
		"templates/helm/Chart.yaml.tmpl":                "helm/Chart.yaml",
		"templates/helm/values.yaml.tmpl":               "helm/values.yaml",
		"templates/helm/templates/deployment.yaml.tmpl": "helm/templates/deployment.yaml",
		"templates/helm/templates/service.yaml.tmpl":    "helm/templates/service.yaml",
		"templates/argocd/application.yaml.tmpl":        "deploy/application.yaml",
	}

	if e.cfg.Spec.Language == "java-spring-boot" {
		mappings["templates/spring-boot/src/main/resources/application.yml.tmpl"] = "src/main/resources/application.yml"

		if e.cfg.Spec.Dependencies.Database != nil {
			mappings["templates/spring-boot/src/test/resources/application.yml.tmpl"] = "src/test/resources/application.yml"
		}
	}

	for tmplPath, outPath := range mappings {
		if err := e.renderTemplate(tmplPath, filepath.Join(outDir, outPath)); err != nil {
			return fmt.Errorf("failed to render %s: %w", tmplPath, err)
		}
	}

	return nil
}

func (e *Engine) renderTemplate(tmplPath, outPath string) error {
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

	if err := tmpl.Execute(f, e.cfg); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
