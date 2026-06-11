package scaffold

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
	for _, m := range GlobalMappings(e.cfg) {
		if err := e.renderTemplate(m.Template, filepath.Join(outDir, m.Output), e.cfg); err != nil {
			return fmt.Errorf("failed to render global %s: %w", m.Template, err)
		}
	}

	for _, svc := range e.cfg.Services {
		svcDir := filepath.Join(outDir, "services", svc.Name)

		type TemplateData struct {
			Config  *config.PlatformConfig
			Service config.Service
		}
		data := TemplateData{
			Config:  e.cfg,
			Service: svc,
		}

		for _, m := range ServiceMappings(e.cfg, svc) {
			if err := e.renderTemplate(m.Template, filepath.Join(svcDir, m.Output), data); err != nil {
				return fmt.Errorf("failed to render %s for %s: %w", m.Template, svc.Name, err)
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
