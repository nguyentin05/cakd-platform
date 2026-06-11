package scaffold

import (
	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/registry"
)

// TemplateMapping represents a single template-to-output file mapping.
type TemplateMapping struct {
	Template string
	Output   string
}

// GlobalMappings returns the list of project-level templates to render based on config.
func GlobalMappings(cfg *config.PlatformConfig) []TemplateMapping {
	mappings := []TemplateMapping{
		{"templates/gitignore.tmpl", ".gitignore"},
	}

	if cfg.Providers.CI != "" {
		mappings = append(mappings, TemplateMapping{"templates/ci/ci.yml.tmpl", ".github/workflows/ci.yml"})
	}

	if cfg.Providers.CD != "" {
		mappings = append(mappings, TemplateMapping{"templates/argocd/application.yaml.tmpl", "deploy/application.yaml"})
	}

	return mappings
}

// ServiceMappings returns the list of per-service templates to render.
func ServiceMappings(cfg *config.PlatformConfig, svc config.Service) []TemplateMapping {
	mappings := []TemplateMapping{
		{"templates/dockerfile.tmpl", "Dockerfile"},
	}

	if cfg.Providers.CD != "" {
		mappings = append(mappings,
			TemplateMapping{"templates/helm/Chart.yaml.tmpl", "helm/Chart.yaml"},
			TemplateMapping{"templates/helm/values.yaml.tmpl", "helm/values.yaml"},
			TemplateMapping{"templates/helm/templates/deployment.yaml.tmpl", "helm/templates/deployment.yaml"},
			TemplateMapping{"templates/helm/templates/service.yaml.tmpl", "helm/templates/service.yaml"},
		)
	}

	if svc.Language == "java-spring-boot" {
		formatExt := "yml"
		if svc.SpringConfigFormat == "properties" {
			formatExt = "properties"
		}

		mappings = append(mappings, TemplateMapping{
			"templates/spring-boot/src/main/resources/application." + formatExt + ".tmpl",
			"src/main/resources/application." + formatExt,
		})

		if usesDB(cfg, svc) {
			mappings = append(mappings, TemplateMapping{
				"templates/spring-boot/src/test/resources/application." + formatExt + ".tmpl",
				"src/test/resources/application." + formatExt,
			})
		}
	}

	return mappings
}

func usesDB(cfg *config.PlatformConfig, svc config.Service) bool {
	for _, use := range svc.Uses {
		for _, b := range cfg.Backing {
			if b.Name == use && (b.Type == registry.PostgreSQL || b.Type == registry.MySQL) {
				return true
			}
		}
	}
	return false
}
