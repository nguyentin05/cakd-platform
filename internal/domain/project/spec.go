package project

const (
	CanonicalAPIVersion = "cakd.io/v1alpha1"
	LegacyAPIVersion    = "platform.dev/v1alpha1"
	KindProject         = "Project"
)

// Project is the aggregate root representing a declarative CAKD platform specification.
type Project struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

// Metadata identifies the project name, owner, and environment.
type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	Owner       string `yaml:"owner" json:"owner"`
	Environment string `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// Spec groups all desired cluster, provider, observability, backing, and service declarations.
type Spec struct {
	Cluster       *ClusterSpec       `yaml:"cluster,omitempty" json:"cluster,omitempty"`
	Providers     *ProvidersSpec     `yaml:"providers,omitempty" json:"providers,omitempty"`
	Observability *ObservabilitySpec `yaml:"observability,omitempty" json:"observability,omitempty"`
	Backing       []BackingSpec      `yaml:"backing,omitempty" json:"backing,omitempty"`
	Services      []ServiceSpec      `yaml:"services" json:"services"`
}

// ClusterSpec defines where and how the target Kubernetes cluster is provisioned.
type ClusterSpec struct {
	Provider     string         `yaml:"provider" json:"provider"`
	Region       string         `yaml:"region,omitempty" json:"region,omitempty"`
	InstanceType string         `yaml:"instanceType,omitempty" json:"instanceType,omitempty"`
	DiskSize     int            `yaml:"diskSize,omitempty" json:"diskSize,omitempty"`
	Domain       string         `yaml:"domain,omitempty" json:"domain,omitempty"`
	Host         string         `yaml:"host,omitempty" json:"host,omitempty"`
	SSHKeyPath   string         `yaml:"sshKeyPath,omitempty" json:"sshKeyPath,omitempty"`
	SSHUser      string         `yaml:"sshUser,omitempty" json:"sshUser,omitempty"`
	Addons       *ClusterAddons `yaml:"addons,omitempty" json:"addons,omitempty"`
}

// ClusterAddons toggles core cluster add-ons installed during cluster initialization.
type ClusterAddons struct {
	ArgoCD     *bool `yaml:"argocd,omitempty" json:"argocd,omitempty"`
	Monitoring *bool `yaml:"monitoring,omitempty" json:"monitoring,omitempty"`
	Logging    *bool `yaml:"logging,omitempty" json:"logging,omitempty"`
	Agent      *bool `yaml:"agent,omitempty" json:"agent,omitempty"`
}

// ProvidersSpec selects concrete tools for VCS, CI/CD, notifications, and observability.
type ProvidersSpec struct {
	VersionControl string `yaml:"versionControl,omitempty" json:"versionControl,omitempty"`
	CI             string `yaml:"ci,omitempty" json:"ci,omitempty"`
	CD             string `yaml:"cd,omitempty" json:"cd,omitempty"`
	Notification   string `yaml:"notification,omitempty" json:"notification,omitempty"`
	LLM            string `yaml:"llm,omitempty" json:"llm,omitempty"`
	Monitoring     string `yaml:"monitoring,omitempty" json:"monitoring,omitempty"`
	Logging        string `yaml:"logging,omitempty" json:"logging,omitempty"`
}

// ObservabilitySpec configures alerting and AI root-cause diagnostics.
type ObservabilitySpec struct {
	Alerting bool      `yaml:"alerting,omitempty" json:"alerting,omitempty"`
	AI       *AIConfig `yaml:"ai,omitempty" json:"ai,omitempty"`
}

// AIConfig configures the LLM model used for AI diagnostics.
type AIConfig struct {
	Model string `yaml:"model" json:"model"`
}

// BackingSpec declares a stateful backing resource (such as PostgreSQL).
type BackingSpec struct {
	Name    string `yaml:"name" json:"name"`
	Type    string `yaml:"type" json:"type"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	Storage string `yaml:"storage,omitempty" json:"storage,omitempty"`
}

// ServiceSpec declares an application microservice and its framework settings.
type ServiceSpec struct {
	Name              string            `yaml:"name" json:"name"`
	Language          string            `yaml:"language" json:"language"`
	Version           string            `yaml:"version,omitempty" json:"version,omitempty"`
	SpringBootVersion string            `yaml:"springBootVersion,omitempty" json:"springBootVersion,omitempty"`
	ProjectBuild      string            `yaml:"projectBuild,omitempty" json:"projectBuild,omitempty"`
	GroupID           string            `yaml:"groupId,omitempty" json:"groupId,omitempty"`
	ArtifactID        string            `yaml:"artifactId,omitempty" json:"artifactId,omitempty"`
	Description       string            `yaml:"description,omitempty" json:"description,omitempty"`
	PackageName       string            `yaml:"packageName,omitempty" json:"packageName,omitempty"`
	Packaging         string            `yaml:"packaging,omitempty" json:"packaging,omitempty"`
	Configuration     string            `yaml:"configuration,omitempty" json:"configuration,omitempty"`
	Dependencies      []string          `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	FrameworkOptions  map[string]string `yaml:"frameworkOptions,omitempty" json:"frameworkOptions,omitempty"`
	Replicas          int               `yaml:"replicas,omitempty" json:"replicas,omitempty"`
	Resources         *ResourcesSpec    `yaml:"resources,omitempty" json:"resources,omitempty"`
	Uses              []string          `yaml:"uses,omitempty" json:"uses,omitempty"`
}

// GetFrameworkOption returns the framework option value for key, checking FrameworkOptions first,
// then mapping to legacy JVM fields (such as "springBootVersion", "groupId", "artifactId",
// "packageName", "packaging"), or returning fallback if not found.
func (s ServiceSpec) GetFrameworkOption(key, fallback string) string {
	if s.FrameworkOptions != nil {
		if val, ok := s.FrameworkOptions[key]; ok && val != "" {
			return val
		}
	}
	switch key {
	case "springBootVersion", "spring_boot_version":
		if s.SpringBootVersion != "" {
			return s.SpringBootVersion
		}
	case "groupId", "group_id":
		if s.GroupID != "" {
			return s.GroupID
		}
	case "artifactId", "artifact_id":
		if s.ArtifactID != "" {
			return s.ArtifactID
		}
	case "packageName", "package_name":
		if s.PackageName != "" {
			return s.PackageName
		}
	case "packaging":
		if s.Packaging != "" {
			return s.Packaging
		}
	case "configuration":
		if s.Configuration != "" {
			return s.Configuration
		}
	case "projectBuild", "project_build":
		if s.ProjectBuild != "" {
			return s.ProjectBuild
		}
	}
	return fallback
}

// ResourcesSpec defines Kubernetes CPU and memory limits for a service.
type ResourcesSpec struct {
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}
