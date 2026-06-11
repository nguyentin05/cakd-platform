package registry

import (
	"github.com/nguyentin05/cakd-platform/internal/registry/backing"
	"github.com/nguyentin05/cakd-platform/internal/registry/providers"
	"github.com/nguyentin05/cakd-platform/internal/registry/services"
	"github.com/nguyentin05/cakd-platform/internal/registry/services/springboot"
)

// Expose constants at the root level for backwards compatibility
const (
	PostgreSQL = backing.PostgreSQL
	MySQL      = backing.MySQL
	Redis      = backing.Redis
	RabbitMQ   = backing.RabbitMQ
	Prometheus = providers.Prometheus
)

// Expose dependency mappings at the root level
var (
	SpringBootDeps = springboot.SpringBootDeps
	MonitoringDeps = providers.MonitoringDeps
)

// SchemaEnums represents the hierarchical structure of valid values
// corresponding to the configuration schema.
type SchemaEnums struct {
	APIVersion []string
	Kind       []string

	Providers ProviderEnums
	Services  ServiceEnums
	Backing   BackingEnums
}

// ProviderEnums holds the valid values for each provider type.
type ProviderEnums struct {
	VersionControl []string
	CI             []string
	CD             []string
	Notification   []string
	LLM            []string
	Monitoring     []string
	Logging        []string
}

// ServiceEnums defines the list of supported service languages, versions, frameworks, build systems, and dependencies.
type ServiceEnums struct {
	Language           []string
	LanguageVersion    []string
	FrameworkVersion   []string
	ProjectBuild       []string
	Packaging          []string
	Dependencies       []string
	SpringConfigFormat []string
}

// BackingEnums specifies the valid options for stateful backing resources.
type BackingEnums struct {
	Type []string
}

// ValidValues is the structured source of truth for all valid enums.
var ValidValues = SchemaEnums{
	APIVersion: []string{"platform.dev/v1alpha1"},
	Kind:       []string{"Project"},
	Providers: ProviderEnums{
		VersionControl: providers.VersionControl,
		CI:             providers.CI,
		CD:             providers.CD,
		Notification:   providers.Notification,
		LLM:            providers.LLM,
		Monitoring:     providers.Monitoring,
		Logging:        providers.Logging,
	},
	Services: ServiceEnums{
		Language:           services.Languages,
		LanguageVersion:    springboot.ValidJavaVersions,
		FrameworkVersion:   springboot.ValidSpringBootVersions,
		ProjectBuild:       springboot.ValidProjectBuilds,
		Packaging:          springboot.ValidPackaging,
		Dependencies:       springboot.ValidSpringDependencies,
		SpringConfigFormat: springboot.ConfigFormats,
	},
	Backing: BackingEnums{
		Type: backing.Types,
	},
}

// Enums is a map representation of all schema enums, mapping tag names to their allowed string values.
var Enums map[string][]string

func init() {
	Enums = map[string][]string{
		"APIVersion":           ValidValues.APIVersion,
		"Kind":                 ValidValues.Kind,
		"vcs":                  ValidValues.Providers.VersionControl,
		"ci":                   ValidValues.Providers.CI,
		"cd":                   ValidValues.Providers.CD,
		"notification":         ValidValues.Providers.Notification,
		"llm":                  ValidValues.Providers.LLM,
		"monitoring":           ValidValues.Providers.Monitoring,
		"logging":              ValidValues.Providers.Logging,
		"language":             ValidValues.Services.Language,
		"java-version":         ValidValues.Services.LanguageVersion,
		"spring-boot-version":  ValidValues.Services.FrameworkVersion,
		"project-build":        ValidValues.Services.ProjectBuild,
		"packaging":            ValidValues.Services.Packaging,
		"spring-dependencies":  ValidValues.Services.Dependencies,
		"spring-config-format": ValidValues.Services.SpringConfigFormat,
		"database":             ValidValues.Backing.Type,
	}
}
