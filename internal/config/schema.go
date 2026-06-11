package config

// PlatformConfig represents the root configuration schema for a CAKD platform project.
// It specifies the services, backing stores, infrastructure providers, and observability settings.
type PlatformConfig struct {
	APIVersion    string         `yaml:"apiVersion" enum:"APIVersion" validate:"required"`
	Kind          string         `yaml:"kind" enum:"Kind" validate:"required"`
	Metadata      Metadata       `yaml:"metadata"`
	Providers     Providers      `yaml:"providers"`
	Services      []Service      `yaml:"services"`
	Backing       []Backing      `yaml:"backing"`
	Observability *Observability `yaml:"observability,omitempty"`
}

// Metadata holds core information about the project including name and owner.
type Metadata struct {
	Name  string `yaml:"name" validate:"required"`
	Owner string `yaml:"owner" validate:"required"`
}

// Providers defines the third-party integrations used for VCS, CI/CD, notifications, LLMs, and monitoring.
type Providers struct {
	VersionControl string `yaml:"versionControl" default:"Providers.VersionControl" enum:"vcs" validate:"required"`
	CI             string `yaml:"ci,omitempty" enum:"ci"`
	CD             string `yaml:"cd,omitempty" enum:"cd"`
	Notification   string `yaml:"notification,omitempty" enum:"notification"`
	LLM            string `yaml:"llm,omitempty" enum:"llm"`
	Monitoring     string `yaml:"monitoring,omitempty" enum:"monitoring"`
	Logging        string `yaml:"logging,omitempty" enum:"logging"`
}

// Service represents a microservice within the platform application, detailing its language, dependencies, and sizing.
type Service struct {
	Name               string     `yaml:"name" validate:"required"`
	Language           string     `yaml:"language" enum:"language" validate:"required"`
	LanguageVersion    string     `yaml:"languageVersion,omitempty" default:"map:LanguageVersion,key:Language" enum:"java-version"`
	FrameworkVersion   string     `yaml:"frameworkVersion,omitempty" enum:"spring-boot-version"`
	ProjectBuild       string     `yaml:"projectBuild,omitempty" enum:"project-build"`
	Packaging          string     `yaml:"packaging,omitempty" enum:"packaging"`
	Dependencies       []string   `yaml:"dependencies,omitempty" enum:"spring-dependencies"`
	SpringConfigFormat string     `yaml:"springConfigFormat,omitempty" enum:"spring-config-format"`
	Replicas           int        `yaml:"replicas,omitempty" default:"Replicas"`
	Resources          *Resources `yaml:"resources,omitempty"`
	Uses               []string   `yaml:"uses,omitempty"`
}

// Resources specifies the CPU and memory limits allocated to a Kubernetes pod.
type Resources struct {
	CPU    string `yaml:"cpu,omitempty" default:"CPU"`
	Memory string `yaml:"memory,omitempty" default:"Memory"`
}

// Backing represents a stateful dependency (like a database or cache) with size and version details.
type Backing struct {
	Name    string `yaml:"name" validate:"required"`
	Type    string `yaml:"type" enum:"database" validate:"required"`
	Version string `yaml:"version,omitempty" default:"map:DBVersion,key:Type"`
	Storage string `yaml:"storage,omitempty" default:"Storage"`
}

// Observability configures alerting rules and AI-assisted diagnostics options.
type Observability struct {
	Alerting bool      `yaml:"alerting,omitempty"`
	AI       *AIConfig `yaml:"ai,omitempty"`
}

// AIConfig defines options for the AI diagnostic engine such as the selected LLM model.
type AIConfig struct {
	Model string `yaml:"model"`
}
