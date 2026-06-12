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

// Enums is a map representation of all schema enums, mapping tag names to their allowed string values.
var Enums = map[string][]string{
	"APIVersion":           {"platform.dev/v1alpha1"},
	"Kind":                 {"Project"},
	"vcs":                  providers.VersionControl,
	"ci":                   providers.CI,
	"cd":                   providers.CD,
	"notification":         providers.Notification,
	"llm":                  providers.LLM,
	"monitoring":           providers.Monitoring,
	"logging":              providers.Logging,
	"language":             services.Languages,
	"java-version":         springboot.ValidJavaVersions,
	"spring-boot-version":  springboot.ValidSpringBootVersions,
	"project-build":        springboot.ValidProjectBuilds,
	"packaging":            springboot.ValidPackaging,
	"spring-dependencies":  springboot.ValidSpringDependencies,
	"spring-config-format": springboot.ConfigFormats,
	"database":             backing.Types,
}
