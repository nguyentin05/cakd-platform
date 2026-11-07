package validate

import (
	"fmt"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
)

// Service validates and applies convention-over-configuration defaults to Project specifications.
type Service struct{}

// NewService creates a new validation Service.
func NewService() *Service {
	return &Service{}
}

// ValidateAndDefault checks structural, enum, cross-field, and DAG rules and populates default values on proj.
//
//nolint:gocyclo
func (s *Service) ValidateAndDefault(proj *project.Project) error {
	if proj == nil {
		return errcode.New(errcode.CodeValidation, "project specification is nil")
	}

	if proj.APIVersion == "" {
		return errcode.New(errcode.CodeValidation, "field 'apiVersion' is required")
	}
	if !catalog.IsAllowed(proj.APIVersion, catalog.APIVersions) {
		return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported apiVersion %q; allowed: %v", proj.APIVersion, catalog.APIVersions))
	}

	if proj.Kind == "" {
		return errcode.New(errcode.CodeValidation, "field 'kind' is required")
	}
	if !catalog.IsAllowed(proj.Kind, catalog.Kinds) {
		return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported kind %q; allowed: %v", proj.Kind, catalog.Kinds))
	}

	if strings.TrimSpace(proj.Metadata.Name) == "" {
		return errcode.New(errcode.CodeValidation, "field 'metadata.name' is required")
	}
	if strings.TrimSpace(proj.Metadata.Owner) == "" {
		return errcode.New(errcode.CodeValidation, "field 'metadata.owner' is required")
	}
	if proj.Metadata.Environment == "" {
		proj.Metadata.Environment = catalog.Defaults.Environment
	}

	applyClusterDefaults(&proj.Spec)
	if err := validateCluster(proj.Spec.Cluster); err != nil {
		return err
	}

	applyProviderDefaults(&proj.Spec)
	if err := validateProviders(proj.Spec.Providers); err != nil {
		return err
	}

	if err := validateObservability(proj.Spec.Observability, proj.Spec.Providers); err != nil {
		return err
	}

	if err := validateAndDefaultBacking(proj.Spec.Backing); err != nil {
		return err
	}

	if err := validateAndDefaultServices(proj); err != nil {
		return err
	}

	return validateDependencyDAG(proj)
}

func applyClusterDefaults(spec *project.Spec) {
	if spec.Cluster == nil {
		spec.Cluster = &project.ClusterSpec{
			Provider: catalog.Defaults.ClusterProvider,
		}
	}
	if spec.Cluster.Provider == "" {
		spec.Cluster.Provider = catalog.Defaults.ClusterProvider
	}
	if spec.Cluster.Region == "" {
		spec.Cluster.Region = catalog.Defaults.Region
	}
	if spec.Cluster.InstanceType == "" {
		spec.Cluster.InstanceType = catalog.Defaults.InstanceType
	}
	if spec.Cluster.DiskSize == 0 {
		spec.Cluster.DiskSize = catalog.Defaults.DiskSize
	}
}

func validateCluster(c *project.ClusterSpec) error {
	if c == nil {
		return nil
	}
	if !catalog.IsAllowed(c.Provider, catalog.ClusterProviders) {
		return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported spec.cluster.provider %q; allowed: %v", c.Provider, catalog.ClusterProviders))
	}
	if c.InstanceType != "" && !catalog.IsAllowed(c.InstanceType, catalog.InstanceTypes) {
		return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported spec.cluster.instanceType %q; allowed: %v", c.InstanceType, catalog.InstanceTypes))
	}
	return nil
}

func applyProviderDefaults(spec *project.Spec) {
	if spec.Providers == nil {
		spec.Providers = &project.ProvidersSpec{}
	}
	if spec.Providers.VersionControl == "" {
		spec.Providers.VersionControl = catalog.Defaults.VersionControl
	}
	if spec.Providers.CI == "" {
		spec.Providers.CI = catalog.Defaults.CI
	}
	if spec.Providers.CD == "" {
		spec.Providers.CD = catalog.Defaults.CD
	}
}

func validateProviders(p *project.ProvidersSpec) error {
	if p == nil {
		return nil
	}
	checks := []struct {
		name    string
		val     string
		allowed []string
	}{
		{"spec.providers.versionControl", p.VersionControl, catalog.VersionControlTypes},
		{"spec.providers.ci", p.CI, catalog.CITypes},
		{"spec.providers.cd", p.CD, catalog.CDTypes},
		{"spec.providers.notification", p.Notification, catalog.NotificationTypes},
		{"spec.providers.llm", p.LLM, catalog.LLMTypes},
		{"spec.providers.monitoring", p.Monitoring, catalog.MonitoringTypes},
		{"spec.providers.logging", p.Logging, catalog.LoggingTypes},
	}
	for _, chk := range checks {
		if chk.val != "" && !catalog.IsAllowed(chk.val, chk.allowed) {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported %s %q; allowed: %v", chk.name, chk.val, chk.allowed))
		}
	}
	if p.CD != "" && p.CI == "" {
		return errcode.New(errcode.CodeValidation, "providers.ci is required when providers.cd is set")
	}
	return nil
}

func validateObservability(obs *project.ObservabilitySpec, p *project.ProvidersSpec) error {
	if obs == nil {
		return nil
	}
	if !obs.Alerting && obs.AI == nil {
		return errcode.New(errcode.CodeValidation, "block 'observability' is declared but empty, must define at least one child")
	}
	if obs.Alerting && (p == nil || p.Notification == "") {
		return errcode.New(errcode.CodeValidation, "providers.notification is required when observability.alerting is enabled")
	}
	if obs.AI != nil {
		if strings.TrimSpace(obs.AI.Model) == "" {
			return errcode.New(errcode.CodeValidation, "observability.ai.model is required when observability.ai is enabled")
		}
		if p == nil || p.LLM == "" {
			return errcode.New(errcode.CodeValidation, "providers.llm is required when observability.ai is enabled")
		}
	}
	return nil
}

func validateAndDefaultBacking(backing []project.BackingSpec) error {
	for i := range backing {
		b := &backing[i]
		if strings.TrimSpace(b.Name) == "" {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("backing[%d].name is required", i))
		}
		if strings.TrimSpace(b.Type) == "" {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("backing[%d].type is required", i))
		}
		if !catalog.IsAllowed(b.Type, catalog.BackingTypes) {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported database %q on backing[%d]; allowed: %v", b.Type, i, catalog.BackingTypes))
		}
		if b.Version == "" {
			b.Version = catalog.Defaults.DBVersion[b.Type]
		}
		if b.Storage == "" {
			b.Storage = catalog.Defaults.Storage
		}
	}
	return nil
}

//nolint:gocyclo
func validateAndDefaultServices(proj *project.Project) error {
	if len(proj.Spec.Services) == 0 {
		return errcode.New(errcode.CodeValidation, "at least one service must be defined in services[]")
	}

	backingTypesByName := make(map[string]string, len(proj.Spec.Backing))
	for _, b := range proj.Spec.Backing {
		backingTypesByName[b.Name] = b.Type
	}
	monitoringProvider := ""
	if proj.Spec.Providers != nil {
		monitoringProvider = proj.Spec.Providers.Monitoring
	}

	for i := range proj.Spec.Services {
		svc := &proj.Spec.Services[i]
		if strings.TrimSpace(svc.Name) == "" {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("services[%d].name is required", i))
		}
		if strings.TrimSpace(svc.Language) == "" {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("services[%d].language is required", i))
		}
		if !catalog.IsAllowed(svc.Language, catalog.ServiceLanguages) {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported language %q on service %q; allowed: %v", svc.Language, svc.Name, catalog.ServiceLanguages))
		}
		if svc.Version == "" {
			svc.Version = catalog.Defaults.LanguageVersion[svc.Language]
		}
		if svc.Replicas == 0 {
			svc.Replicas = catalog.Defaults.Replicas
		}
		if svc.Resources == nil {
			svc.Resources = &project.ResourcesSpec{}
		}
		if svc.Resources.CPU == "" {
			svc.Resources.CPU = catalog.Defaults.CPU
		}
		if svc.Resources.Memory == "" {
			svc.Resources.Memory = catalog.Defaults.Memory
		}

		if svc.Language == catalog.LangJavaSpringBoot {
			if !catalog.IsAllowed(svc.Version, catalog.ValidJavaVersions) {
				return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported java-version %q on service %q; allowed: %v", svc.Version, svc.Name, catalog.ValidJavaVersions))
			}
			if svc.SpringBootVersion != "" && !catalog.IsAllowed(svc.SpringBootVersion, catalog.ValidSpringBootVersions) {
				return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported spring-boot-version %q on service %q", svc.SpringBootVersion, svc.Name))
			}
			if svc.ProjectBuild != "" && !catalog.IsAllowed(svc.ProjectBuild, catalog.ValidProjectBuilds) {
				return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported project-build %q on service %q", svc.ProjectBuild, svc.Name))
			}
			if svc.Packaging != "" && !catalog.IsAllowed(svc.Packaging, catalog.ValidPackaging) {
				return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported packaging %q on service %q", svc.Packaging, svc.Name))
			}
			if svc.Configuration != "" && !catalog.IsAllowed(svc.Configuration, catalog.SpringConfigFormats) {
				return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported spring-config-format %q on service %q", svc.Configuration, svc.Name))
			}
			for _, dep := range svc.Dependencies {
				if !catalog.IsAllowed(dep, catalog.ValidSpringDependencies) {
					return errcode.New(errcode.CodeValidation, fmt.Sprintf("unsupported spring-dependencies %q on service %q", dep, svc.Name))
				}
			}

			var usedBackingTypes []string
			for _, u := range svc.Uses {
				if bt, ok := backingTypesByName[u]; ok {
					usedBackingTypes = append(usedBackingTypes, bt)
				}
			}
			svc.Dependencies = catalog.ResolveSpringBootDependencies(svc.Dependencies, usedBackingTypes, monitoringProvider)
		}
	}
	return nil
}

func validateDependencyDAG(proj *project.Project) error {
	backingNames := make(map[string]bool, len(proj.Spec.Backing))
	for _, b := range proj.Spec.Backing {
		backingNames[b.Name] = true
	}

	serviceEdges := make(map[string][]string, len(proj.Spec.Services))
	for _, svc := range proj.Spec.Services {
		serviceEdges[svc.Name] = nil
	}

	for _, svc := range proj.Spec.Services {
		for _, target := range svc.Uses {
			if backingNames[target] {
				continue
			}
			if _, isSvc := serviceEdges[target]; isSvc {
				serviceEdges[svc.Name] = append(serviceEdges[svc.Name], target)
				continue
			}
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("service %q references unknown backing resource %q", svc.Name, target))
		}
	}

	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)
	state := make(map[string]int, len(serviceEdges))
	var visit func(node string) error
	visit = func(node string) error {
		state[node] = visiting
		for _, next := range serviceEdges[node] {
			if state[next] == visiting {
				return errcode.New(errcode.CodeValidation, fmt.Sprintf("cyclic dependency detected between services %q and %q", node, next))
			}
			if state[next] == unvisited {
				if err := visit(next); err != nil {
					return err
				}
			}
		}
		state[node] = visited
		return nil
	}

	for _, svc := range proj.Spec.Services {
		if state[svc.Name] == unvisited {
			if err := visit(svc.Name); err != nil {
				return err
			}
		}
	}
	return nil
}
