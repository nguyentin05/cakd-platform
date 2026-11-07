package catalog

import "slices"

const (
	BackingPostgreSQL = "postgresql"
	BackingMySQL      = "mysql"
	BackingRedis      = "redis"
	BackingRabbitMQ   = "rabbitmq"

	ProviderGitHub        = "github"
	ProviderGitHubActions = "github-actions"
	ProviderArgoCD        = "argocd"
	ProviderDiscord       = "discord"
	ProviderGemini        = "gemini"
	ProviderPrometheus    = "prometheus"
	ProviderLoki          = "loki"

	ClusterLocal  = "local"
	ClusterAWSEC2 = "aws-ec2"

	LangJavaSpringBoot = "java-spring-boot"
)

// PlatformDefaults holds convention-over-configuration defaults for project specifications.
type PlatformDefaults struct {
	LanguageVersion map[string]string
	DBVersion       map[string]string
	Storage         string
	Replicas        int
	CPU             string
	Memory          string
	VersionControl  string
	CI              string
	CD              string
	ClusterProvider string
	Region          string
	InstanceType    string
	DiskSize        int
	Environment     string
}

// Defaults provides the single source of truth for implicit configuration values.
var Defaults = PlatformDefaults{
	LanguageVersion: map[string]string{
		LangJavaSpringBoot: "21",
	},
	DBVersion: map[string]string{
		BackingPostgreSQL: "16",
	},
	Storage:         "5Gi",
	Replicas:        1,
	CPU:             "500m",
	Memory:          "512Mi",
	VersionControl:  ProviderGitHub,
	CI:              ProviderGitHubActions,
	CD:              ProviderArgoCD,
	ClusterProvider: ClusterLocal,
	Region:          "ap-southeast-1",
	InstanceType:    "t3.medium",
	DiskSize:        30,
	Environment:     "production",
}

var (
	APIVersions          = []string{"cakd.io/v1alpha1", "platform.dev/v1alpha1"}
	Kinds                = []string{"Project"}
	ClusterProviders     = []string{ClusterLocal, ClusterAWSEC2}
	InstanceTypes        = []string{"t3.micro", "t3.small", "t3.medium", "t3.large"}
	VersionControlTypes  = []string{ProviderGitHub}
	CITypes              = []string{ProviderGitHubActions}
	CDTypes              = []string{ProviderArgoCD}
	NotificationTypes    = []string{ProviderDiscord}
	LLMTypes             = []string{ProviderGemini}
	MonitoringTypes      = []string{ProviderPrometheus}
	LoggingTypes         = []string{ProviderLoki}
	ServiceLanguages     = []string{LangJavaSpringBoot}
	BackingTypes         = []string{BackingPostgreSQL}
	SpringConfigFormats  = []string{"yaml", "properties"}
	SpringBackingDepsMap = map[string][]string{
		BackingPostgreSQL: {"data-jpa", "postgresql"},
		BackingMySQL:      {"data-jpa", "mysql"},
		BackingRedis:      {"data-redis"},
		BackingRabbitMQ:   {"amqp"},
	}
	MonitoringDepsMap = map[string]string{
		ProviderPrometheus: "prometheus",
	}
)

// ResolveSpringBootDependencies merges user dependencies with backing and monitoring dependencies.
func ResolveSpringBootDependencies(userDeps, backingTypes []string, monitoringProvider string) []string {
	seen := make(map[string]bool, len(userDeps)+4)
	result := make([]string, 0, len(userDeps)+4)
	add := func(dep string) {
		if dep != "" && !seen[dep] {
			seen[dep] = true
			result = append(result, dep)
		}
	}

	for _, d := range userDeps {
		add(d)
	}
	for _, bt := range backingTypes {
		for _, dep := range SpringBackingDepsMap[bt] {
			add(dep)
		}
	}
	if monDep, ok := MonitoringDepsMap[monitoringProvider]; ok {
		add("actuator")
		add(monDep)
	}
	return result
}

// IsAllowed reports whether val is present in the allowed slice.
func IsAllowed(val string, allowed []string) bool {
	return slices.Contains(allowed, val)
}
