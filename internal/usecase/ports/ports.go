package ports

import (
	"context"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
)

// SpecRepository loads, normalizes, and persists Project specifications.
type SpecRepository interface {
	LoadFile(path string) (*project.Project, error)
	LoadBytes(data []byte) (*project.Project, error)
	SaveFile(path string, proj *project.Project) error
	MarshalCanonical(proj *project.Project) ([]byte, error)
}

// CredentialSet holds resolved credentials for external providers.
type CredentialSet struct {
	GitHubToken     string `json:"githubToken,omitempty"`
	DiscordBotToken string `json:"discordBotToken,omitempty"`
	DiscordGuildID  string `json:"discordGuildId,omitempty"`
	GeminiAPIKey    string `json:"geminiApiKey,omitempty"`
}

// CredentialStatus summarizes configured credentials with masked values.
type CredentialStatus struct {
	GitHubToken     string `json:"githubToken"`
	DiscordBotToken string `json:"discordBotToken"`
	DiscordGuildID  string `json:"discordGuildId"`
	GeminiAPIKey    string `json:"geminiApiKey"`
}

// CredentialStore loads and persists non-interactive credentials.
type CredentialStore interface {
	Get() (CredentialSet, error)
	Save(creds CredentialSet) error
	Delete(provider string) error
	Status() (CredentialStatus, error)
	Path() string
}

// ClusterState captures persisted metadata from an initialized cluster.
type ClusterState struct {
	Provider       string `json:"provider" yaml:"provider"`
	Host           string `json:"host,omitempty" yaml:"host,omitempty"`
	Domain         string `json:"domain,omitempty" yaml:"domain,omitempty"`
	KubeconfigPath string `json:"kubeconfigPath,omitempty" yaml:"kubeconfigPath,omitempty"`
	SSHKeyPath     string `json:"sshKeyPath,omitempty" yaml:"sshKeyPath,omitempty"`
	SSHUser        string `json:"sshUser,omitempty" yaml:"sshUser,omitempty"`
}

// ClusterStateStore persists cluster state and Alertmanager webhook routing maps.
type ClusterStateStore interface {
	SaveClusterState(state ClusterState) error
	LoadClusterState() (ClusterState, error)
	SaveWebhook(ctx context.Context, projectName, webhookURL string) error
	LoadWebhooks() (map[string]string, error)
}

// CloudHostProvisionRequest configures cloud host provisioning via Terraform.
type CloudHostProvisionRequest struct {
	ProjectName  string
	Region       string
	InstanceType string
	DiskSize     int
	WorkDir      string
}

// CloudHostProvisionResult contains the provisioned host IP and generated SSH key path.
type CloudHostProvisionResult struct {
	PublicIP   string
	SSHKeyPath string
}

// CloudHostProvisioner provisions and destroys cloud compute instances.
type CloudHostProvisioner interface {
	ProvisionHost(ctx context.Context, req CloudHostProvisionRequest) (CloudHostProvisionResult, error)
	DestroyHost(ctx context.Context, workDir string) error
}

// ClusterBootstrapOptions configures Kubernetes cluster initialization and add-ons.
type ClusterBootstrapOptions struct {
	Provider       string
	Host           string
	SSHKeyPath     string
	SSHUser        string
	Domain         string
	InstallArgoCD  bool
	InstallMonitor bool
	InstallLogging bool
	InstallAgent   bool
	GeminiAPIKey   string
	GeminiModel    string
	DiscordWebhook string
}

// ClusterBootstrapper initializes K3s/local clusters and installs core platform workloads.
type ClusterBootstrapper interface {
	BootstrapCluster(ctx context.Context, opts ClusterBootstrapOptions) (ClusterState, error)
}

// ScaffoldRequest specifies the input for project and service template rendering.
type ScaffoldRequest struct {
	Project    *project.Project
	BaseDomain string
	OutputDir  string
}

// TemplateRenderer generates service source code, Helm charts, CI workflows, and GitOps manifests.
type TemplateRenderer interface {
	RenderProject(ctx context.Context, req ScaffoldRequest) error
}

// FrameworkGenerator downloads or generates framework starter archives (such as Spring Initializr).
type FrameworkGenerator interface {
	GenerateService(ctx context.Context, svc project.ServiceSpec, destDir string) error
}

// RepoProvisionRequest configures remote VCS repository creation and initial push.
type RepoProvisionRequest struct {
	ProjectName string
	Owner       string
	Token       string
	SourceDir   string
	StateDir    string
}

// VCSProvider provisions Git repositories, pushes scaffolded code, and supports rollback.
type VCSProvider interface {
	ProvisionAndPush(ctx context.Context, req RepoProvisionRequest) (repoURL string, err error)
	Rollback(ctx context.Context, req RepoProvisionRequest) error
}

// GitOpsRegisterRequest configures GitOps application registration.
type GitOpsRegisterRequest struct {
	AppName        string
	RepoURL        string
	TargetPath     string
	Namespace      string
	KubeconfigPath string
}

// GitOpsProvider registers and syncs applications in the cluster CD controller.
type GitOpsProvider interface {
	RegisterApplication(ctx context.Context, req GitOpsRegisterRequest) error
}

// NotificationProvider manages alert channels and sends structured notifications.
type NotificationProvider interface {
	ProvisionWebhook(ctx context.Context, channelName string) (webhookURL string, err error)
	SendAlert(ctx context.Context, webhookURL string, items []diagnostic.NotificationItem) error
	SendMessage(ctx context.Context, webhookURL, content string) error
}

// MetricsCollector queries cluster metrics for diagnostic summaries.
type MetricsCollector interface {
	CollectMetricsSummary(ctx context.Context, namespace string) (string, error)
}

// LogCollector queries cluster logs for diagnostic summaries.
type LogCollector interface {
	CollectLogsSummary(ctx context.Context, namespace string) (string, error)
}

// AIAnalyzer performs LLM-based root-cause analysis on metrics, logs, or alerts.
type AIAnalyzer interface {
	Analyze(ctx context.Context, prompt string) (string, error)
}
