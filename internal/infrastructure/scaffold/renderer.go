package scaffold

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

//go:embed templates
var templateFS embed.FS

// Renderer implements ports.TemplateRenderer using embedded [[ ]] delimited templates.
type Renderer struct{}

// NewRenderer creates a new scaffold template Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

type templateMapping struct {
	templatePath string
	outputPath   string
}

// serviceView provides template-compatible fields for both new and existing templates.
type serviceView struct {
	Name               string
	Language           string
	Version            string
	LanguageVersion    string
	SpringBootVersion  string
	FrameworkVersion   string
	ProjectBuild       string
	Packaging          string
	SpringConfigFormat string
	Dependencies       []string
	Replicas           int
	Resources          project.ResourcesSpec
	Uses               []string
}

type globalView struct {
	Metadata project.Metadata
	Spec     project.Spec
	Services []serviceView
}

type serviceRenderContext struct {
	Config  globalView
	Service serviceView
	Domain  string
}

// RenderProject renders all global and per-service templates into req.OutputDir.
func (r *Renderer) RenderProject(_ context.Context, req ports.ScaffoldRequest) error {
	if req.Project == nil {
		return errcode.New(errcode.CodeValidation, "project is required for template rendering")
	}
	domain := req.BaseDomain
	if domain == "" {
		domain = "127.0.0.1.nip.io"
	}

	gView := toGlobalView(req.Project)

	for _, m := range globalMappings(req.Project) {
		dest := filepath.Join(req.OutputDir, m.outputPath)
		if err := renderFile(m.templatePath, dest, gView); err != nil {
			return err
		}
	}

	for i, svc := range req.Project.Spec.Services {
		svcDir := filepath.Join(req.OutputDir, "services", svc.Name)
		ctxData := serviceRenderContext{
			Config:  gView,
			Service: gView.Services[i],
			Domain:  domain,
		}
		for _, m := range serviceMappings(req.Project, svc) {
			dest := filepath.Join(svcDir, m.outputPath)
			if err := renderFile(m.templatePath, dest, ctxData); err != nil {
				return err
			}
		}
	}

	return nil
}

func toGlobalView(proj *project.Project) globalView {
	svcs := make([]serviceView, 0, len(proj.Spec.Services))
	for _, s := range proj.Spec.Services {
		res := project.ResourcesSpec{
			CPU:    catalog.Defaults.CPU,
			Memory: catalog.Defaults.Memory,
		}
		if s.Resources != nil {
			if s.Resources.CPU != "" {
				res.CPU = s.Resources.CPU
			}
			if s.Resources.Memory != "" {
				res.Memory = s.Resources.Memory
			}
		}
		ver := s.Version
		if ver == "" {
			ver = catalog.Defaults.LanguageVersion[s.Language]
		}
		svcs = append(svcs, serviceView{
			Name:               s.Name,
			Language:           s.Language,
			Version:            ver,
			LanguageVersion:    ver,
			SpringBootVersion:  s.SpringBootVersion,
			FrameworkVersion:   s.SpringBootVersion,
			ProjectBuild:       s.ProjectBuild,
			Packaging:          s.Packaging,
			SpringConfigFormat: s.Configuration,
			Dependencies:       s.Dependencies,
			Replicas:           s.Replicas,
			Resources:          res,
			Uses:               s.Uses,
		})
	}
	return globalView{
		Metadata: proj.Metadata,
		Spec:     proj.Spec,
		Services: svcs,
	}
}

func globalMappings(proj *project.Project) []templateMapping {
	mappings := []templateMapping{
		{"templates/gitignore.tmpl", ".gitignore"},
	}
	if proj.Spec.Providers != nil {
		if proj.Spec.Providers.CI != "" {
			mappings = append(mappings, templateMapping{"templates/ci/ci.yml.tmpl", ".github/workflows/ci.yml"})
		}
		if proj.Spec.Providers.CD != "" {
			mappings = append(mappings, templateMapping{"templates/argocd/application.yaml.tmpl", "deploy/application.yaml"})
		}
	}
	return mappings
}

func serviceMappings(proj *project.Project, svc project.ServiceSpec) []templateMapping {
	mappings := []templateMapping{
		{"templates/dockerfile.tmpl", "Dockerfile"},
	}
	if proj.Spec.Providers != nil && proj.Spec.Providers.CD != "" {
		mappings = append(mappings,
			templateMapping{"templates/helm/Chart.yaml.tmpl", "helm/Chart.yaml"},
			templateMapping{"templates/helm/values.yaml.tmpl", "helm/values.yaml"},
			templateMapping{"templates/helm/templates/deployment.yaml.tmpl", "helm/templates/deployment.yaml"},
			templateMapping{"templates/helm/templates/service.yaml.tmpl", "helm/templates/service.yaml"},
			templateMapping{"templates/helm/templates/ingress.yaml.tmpl", "helm/templates/ingress.yaml"},
		)
	}
	if svc.Language == catalog.LangJavaSpringBoot {
		ext := "yml"
		if svc.Configuration == "properties" {
			ext = "properties"
		}
		mappings = append(mappings, templateMapping{
			"templates/spring-boot/src/main/resources/application." + ext + ".tmpl",
			"src/main/resources/application." + ext,
		})
		if usesRelationalDB(proj, svc) {
			mappings = append(mappings, templateMapping{
				"templates/spring-boot/src/test/resources/application." + ext + ".tmpl",
				"src/test/resources/application." + ext,
			})
		}
	}
	return mappings
}

func usesRelationalDB(proj *project.Project, svc project.ServiceSpec) bool {
	for _, u := range svc.Uses {
		for _, b := range proj.Spec.Backing {
			if b.Name == u && (b.Type == catalog.BackingPostgreSQL || b.Type == catalog.BackingMySQL) {
				return true
			}
		}
	}
	return false
}

func renderFile(tmplPath, outPath string, data any) error {
	raw, err := templateFS.ReadFile(tmplPath)
	if err != nil {
		return errcode.Wrap(errcode.CodeProvision, "read embedded template "+tmplPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o750); err != nil {
		return errcode.Wrap(errcode.CodeProvision, "create template directory", err)
	}
	tmpl, err := template.New(filepath.Base(tmplPath)).Delims("[[", "]]").Parse(string(raw))
	if err != nil {
		return errcode.Wrap(errcode.CodeProvision, "parse template "+tmplPath, err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return errcode.Wrap(errcode.CodeProvision, "create output file "+outPath, err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, data); err != nil {
		return errcode.Wrap(errcode.CodeProvision, "execute template "+tmplPath, err)
	}
	return nil
}
