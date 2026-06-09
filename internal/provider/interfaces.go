package provider

import "github.com/nguyentin05/cakd-platform/internal/config"

type Notifier interface {
	ProvisionChannel(projectName string) (webhookURL string, err error)
	SendAlert(webhookURL string, payload AlertPayload) error
}

type AlertPayload struct {
	Items []AlertItem
}

type AlertItem struct {
	Title       string
	Description string
	Severity    string
}

type LLM interface {
	Analyze(context string) (diagnosis string, err error)
}

type VCS interface {
	InitAndPush(dir string, repoURL string, token string) error
}

type AppFramework interface {
	Scaffold(cfg *config.PlatformConfig, svc config.Service, outDir string) error
}

type IaC interface {
	Apply() (map[string]string, error)
	Destroy() error
}
