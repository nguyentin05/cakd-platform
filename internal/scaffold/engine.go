package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nguyentin05/cakd-platform/internal/schema"
)

//go:embed templates
var content embed.FS

// Engine manages loading, parsing, and rendering embedded template files into the target project output directory.
type Engine struct {
	cfg *schema.PlatformConfig
}

// TemplateData holds the required context for rendering scaffold templates.
// It is exposed so templates can be tested independently of the Generate execution.
type TemplateData struct {
	Config  *schema.PlatformConfig
	Service schema.Service
}

// New initializes and returns a new template rendering Engine.
func New(cfg *schema.PlatformConfig) *Engine {
	return &Engine{cfg: cfg}
}

// Generate iterates over all global and service-specific template mappings, rendering
// and writing them to their respective locations in the output directory.
func (e *Engine) Generate(outDir string) error {
	for _, m := range GlobalMappings(e.cfg) {
		if err := e.renderTemplate(m.Template, filepath.Join(outDir, m.Output), e.cfg); err != nil {
			return fmt.Errorf("failed to render global %s: %w", m.Template, err)
		}
	}

	for _, svc := range e.cfg.Services {
		svcDir := filepath.Join(outDir, "services", svc.Name)

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

// renderTemplate reads an embedded template file, parses it using double square brackets ("[[", "]]")
// as delimiters to avoid conflict with standard Helm/Go brackets, and executes it to the outPath.
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
