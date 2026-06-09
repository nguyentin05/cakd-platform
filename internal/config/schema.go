package config

type PlatformConfig struct {
	APIVersion    string         `yaml:"apiVersion" enum:"APIVersion" validate:"required"`
	Kind          string         `yaml:"kind" enum:"Kind" validate:"required"`
	Metadata      Metadata       `yaml:"metadata"`
	Providers     Providers      `yaml:"providers"`
	Services      []Service      `yaml:"services"`
	Backing       []Backing      `yaml:"backing"`
	Observability *Observability `yaml:"observability,omitempty"`
}

type Metadata struct {
	Name  string `yaml:"name" validate:"required"`
	Owner string `yaml:"owner" validate:"required"`
}

type Providers struct {
	VersionControl string `yaml:"versionControl" default:"Providers.VersionControl" enum:"vcs" validate:"required"`
	CI             string `yaml:"ci,omitempty" enum:"ci"`
	CD             string `yaml:"cd,omitempty" enum:"cd"`
	Notification   string `yaml:"notification,omitempty" enum:"notification"`
	LLM            string `yaml:"llm,omitempty" enum:"llm"`
	Monitoring     string `yaml:"monitoring,omitempty" enum:"monitoring"`
	Logging        string `yaml:"logging,omitempty" enum:"logging"`
}

type Service struct {
	Name      string     `yaml:"name" validate:"required"`
	Language  string     `yaml:"language" enum:"language" validate:"required"`
	Version   string     `yaml:"version,omitempty" default:"map:LanguageVersion,key:Language"`
	Replicas  int        `yaml:"replicas,omitempty" default:"Replicas"`
	Resources *Resources `yaml:"resources,omitempty"`
	Uses      []string   `yaml:"uses,omitempty"`
}

type Resources struct {
	CPU    string `yaml:"cpu,omitempty" default:"CPU"`
	Memory string `yaml:"memory,omitempty" default:"Memory"`
}

type Backing struct {
	Name    string `yaml:"name" validate:"required"`
	Type    string `yaml:"type" enum:"database" validate:"required"`
	Version string `yaml:"version,omitempty" default:"map:DBVersion,key:Type"`
	Storage string `yaml:"storage,omitempty" default:"Storage"`
}

type Observability struct {
	Alerting bool      `yaml:"alerting,omitempty"`
	AI       *AIConfig `yaml:"ai,omitempty"`
}

type AIConfig struct {
	Model string `yaml:"model"`
}
