package registry

// Enums acts as the Single Source of Truth (SSOT) for all valid values
// allowed in the system configuration schema.
var Enums = map[string][]string{
	"APIVersion":   {"platform.dev/v1alpha1"},
	"Kind":         {"Project"},
	"vcs":          {"github"},
	"ci":           {"github-actions"},
	"cd":           {"argocd"},
	"notification": {"discord"},
	"llm":          {"gemini"},
	"monitoring":   {Prometheus},
	"logging":      {"loki"},
	"language":     {"java-spring-boot"},
	"database":     {PostgreSQL},
}
