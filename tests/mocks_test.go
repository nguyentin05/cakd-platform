package tests_test

import (
	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/iac"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
)

// MockAppFramework implements app.AppFramework
type MockAppFramework struct {
	ScaffoldFunc func(cfg *config.PlatformConfig, svc config.Service, outDir string) error
	Scaffolded   bool
}

func (m *MockAppFramework) Scaffold(cfg *config.PlatformConfig, svc config.Service, outDir string) error {
	m.Scaffolded = true
	if m.ScaffoldFunc != nil {
		return m.ScaffoldFunc(cfg, svc, outDir)
	}
	return nil
}

// MockVersionControl implements version_control.VersionControl
type MockVersionControl struct {
	InitAndPushFunc func(dir string, repoURL string, token string) error
	Pushed          bool
}

func (m *MockVersionControl) InitAndPush(dir string, repoURL string, token string) error {
	m.Pushed = true
	if m.InitAndPushFunc != nil {
		return m.InitAndPushFunc(dir, repoURL, token)
	}
	return nil
}

// MockCD implements cd.CD
type MockCD struct {
	RegisterFunc func(manifestPath string) error
	Registered   bool
}

func (m *MockCD) Register(manifestPath string) error {
	m.Registered = true
	if m.RegisterFunc != nil {
		return m.RegisterFunc(manifestPath)
	}
	return nil
}

// MockNotifier implements notify.Notifier
type MockNotifier struct {
	ProvisionChannelFunc func(projectName string) (webhookURL string, err error)
	SendAlertFunc        func(webhookURL string, payload notify.AlertPayload) error
	Provisioned          bool
}

func (m *MockNotifier) ProvisionChannel(projectName string) (webhookURL string, err error) {
	m.Provisioned = true
	if m.ProvisionChannelFunc != nil {
		return m.ProvisionChannelFunc(projectName)
	}
	return "http://mock-webhook.discord/" + projectName, nil
}

func (m *MockNotifier) SendAlert(webhookURL string, payload notify.AlertPayload) error {
	if m.SendAlertFunc != nil {
		return m.SendAlertFunc(webhookURL, payload)
	}
	return nil
}

// MockIaCEngine implements iac.Engine
type MockIaCEngine struct {
	ApplyFunc   func() (map[string]string, error)
	DestroyFunc func() error
	Applied     bool
	Destroyed   bool
}

func (m *MockIaCEngine) Apply() (map[string]string, error) {
	m.Applied = true
	if m.ApplyFunc != nil {
		return m.ApplyFunc()
	}
	return map[string]string{
		"repo_clone_url": "https://github.com/tin/demo.git",
		"repo_html_url":  "https://github.com/tin/demo",
	}, nil
}

func (m *MockIaCEngine) Destroy() error {
	m.Destroyed = true
	if m.DestroyFunc != nil {
		return m.DestroyFunc()
	}
	return nil
}

// SetupMockIaC returns a function that overrides iac.NewEngine factory
func SetupMockIaC(mock *MockIaCEngine) func() {
	orig := iac.NewEngine
	iac.NewEngine = func(cfg *config.PlatformConfig, workDir string) iac.Engine {
		return mock
	}
	return func() {
		iac.NewEngine = orig
	}
}
