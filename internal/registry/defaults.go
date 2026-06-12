package registry

// PlatformDefaults defines the structured defaults for CAKD platform configuration.
type PlatformDefaults struct {
	LanguageVersion map[string]string
	DBVersion       map[string]string
	Storage         string
	Replicas        int
	CPU             string
	Memory          string
	Providers       ProviderDefaults
}

// ProviderDefaults defines the defaults for VCS and CI/CD providers.
type ProviderDefaults struct {
	VersionControl string
	CI             string
	CD             string
}

// Defaults serves as the Single Source of Truth (SSOT) for all implicit platform configurations.
// Adhering to the "Convention over Configuration" paradigm, it ensures that users
// are only required to declare explicit deviations from these baseline definitions.
//
// When extending CAKD to support new technologies, their default versions and
// resource allocations must be registered here.
var Defaults = PlatformDefaults{
	LanguageVersion: map[string]string{
		"java-spring-boot": "21",
	},
	DBVersion: map[string]string{
		PostgreSQL: "16",
	},
	Storage:  "5Gi",
	Replicas: 1,
	CPU:      "500m",
	Memory:   "512Mi",
	Providers: ProviderDefaults{
		VersionControl: "github",
		CI:             "github-actions",
		CD:             "argocd",
	},
}
