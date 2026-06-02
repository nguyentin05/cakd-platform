package config

type PlatformConfig struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name  string `yaml:"name"`
	Owner string `yaml:"owner"`
}

type Spec struct {
	Language     string       `yaml:"language"`
	Version      string       `yaml:"version"`
	Features     Features     `yaml:"features"`
	Dependencies Dependencies `yaml:"dependencies"`
}

type Features struct {
	Monitoring *bool `yaml:"monitoring,omitempty"`
	Alerting   *bool `yaml:"alerting,omitempty"`
}

type Dependencies struct {
	Database *Database `yaml:"database,omitempty"`
}

type Database struct {
	Type    string `yaml:"type"`
	Version string `yaml:"version"`
	Storage string `yaml:"storage"`
}
