package tests_test

import (
	"errors"
	"os"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config/defaults"
	"github.com/nguyentin05/cakd-platform/internal/pipeline"
	"github.com/nguyentin05/cakd-platform/internal/provider"
	"github.com/nguyentin05/cakd-platform/internal/provider/app"
	"github.com/nguyentin05/cakd-platform/internal/provider/cd"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
	"github.com/nguyentin05/cakd-platform/internal/provider/version_control"
	"github.com/nguyentin05/cakd-platform/internal/schema"
)

const (
	testAPIVersion      = "platform.dev/v1alpha1"
	testProjectKind     = "Project"
	testOwnerTin        = "tin"
	testVcsGithub       = "github"
	testCdArgoCD        = "argocd"
	testSvcAPI          = "api"
	testLangSpringBoot  = "java-spring-boot"
	testNotifierDiscord = "discord"
)

func TestExecute_Success(t *testing.T) {
	tempDir := t.TempDir()

	mockApp := &MockAppFramework{}
	mockVCS := &MockVersionControl{}
	mockCD := &MockCD{}
	mockNotifier := &MockNotifier{}
	mockIaC := &MockIaCEngine{}

	// Setup mock IaC
	restoreIaC := SetupMockIaC(mockIaC)
	defer restoreIaC()

	// Backup and override global provider registry
	origApp := provider.Providers.AppFrameworks[testLangSpringBoot]
	origVCS := provider.Providers.VersionControls[testVcsGithub]
	origCD := provider.Providers.CDs[testCdArgoCD]
	origNotifier := provider.Providers.Notifiers[testNotifierDiscord]

	provider.Providers.AppFrameworks[testLangSpringBoot] = func() app.AppFramework { return mockApp }
	provider.Providers.VersionControls[testVcsGithub] = func() version_control.VersionControl { return mockVCS }
	provider.Providers.CDs[testCdArgoCD] = func() cd.CD { return mockCD }
	provider.Providers.Notifiers[testNotifierDiscord] = func() notify.Notifier { return mockNotifier }

	defer func() {
		provider.Providers.AppFrameworks[testLangSpringBoot] = origApp
		provider.Providers.VersionControls[testVcsGithub] = origVCS
		provider.Providers.CDs[testCdArgoCD] = origCD
		provider.Providers.Notifiers[testNotifierDiscord] = origNotifier
	}()

	cfg := &schema.PlatformConfig{
		APIVersion: testAPIVersion,
		Kind:       testProjectKind,
		Metadata: schema.Metadata{
			Name:  "test-success-project",
			Owner: testOwnerTin,
		},
		Providers: schema.Providers{
			VersionControl: testVcsGithub,
			CI:             "github-actions",
			CD:             testCdArgoCD,
			Notification:   testNotifierDiscord,
		},
		Services: []schema.Service{
			{
				Name:            testSvcAPI,
				Language:        testLangSpringBoot,
				LanguageVersion: "21",
			},
		},
	}

	defaults.Apply(cfg)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

	err = pipeline.Execute(cfg, true)
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Verify all steps executed
	if !mockApp.Scaffolded {
		t.Error("Expected app framework scaffolding to run")
	}
	if !mockIaC.Applied {
		t.Error("Expected IaC apply to run")
	}
	if !mockNotifier.Provisioned {
		t.Error("Expected Notification provisioning to run")
	}
	if !mockVCS.Pushed {
		t.Error("Expected Git push to repository to run")
	}
	if !mockCD.Registered {
		t.Error("Expected CD registration to run")
	}
	if mockIaC.Destroyed {
		t.Error("Did not expect IaC rollback/destroy to be called on success")
	}
}

func TestExecute_RollbackOnGitPushFailure(t *testing.T) {
	tempDir := t.TempDir()

	mockApp := &MockAppFramework{}
	mockVCS := &MockVersionControl{
		InitAndPushFunc: func(dir, repoURL, token string) error {
			return errors.New("simulated git push error")
		},
	}
	mockCD := &MockCD{}
	mockNotifier := &MockNotifier{}
	mockIaC := &MockIaCEngine{}

	restoreIaC := SetupMockIaC(mockIaC)
	defer restoreIaC()

	origApp := provider.Providers.AppFrameworks[testLangSpringBoot]
	origVCS := provider.Providers.VersionControls[testVcsGithub]
	origCD := provider.Providers.CDs[testCdArgoCD]
	origNotifier := provider.Providers.Notifiers[testNotifierDiscord]

	provider.Providers.AppFrameworks[testLangSpringBoot] = func() app.AppFramework { return mockApp }
	provider.Providers.VersionControls[testVcsGithub] = func() version_control.VersionControl { return mockVCS }
	provider.Providers.CDs[testCdArgoCD] = func() cd.CD { return mockCD }
	provider.Providers.Notifiers[testNotifierDiscord] = func() notify.Notifier { return mockNotifier }

	defer func() {
		provider.Providers.AppFrameworks[testLangSpringBoot] = origApp
		provider.Providers.VersionControls[testVcsGithub] = origVCS
		provider.Providers.CDs[testCdArgoCD] = origCD
		provider.Providers.Notifiers[testNotifierDiscord] = origNotifier
	}()

	cfg := &schema.PlatformConfig{
		APIVersion: testAPIVersion,
		Kind:       testProjectKind,
		Metadata: schema.Metadata{
			Name:  "test-rollback-project",
			Owner: testOwnerTin,
		},
		Providers: schema.Providers{
			VersionControl: testVcsGithub,
			CI:             "github-actions",
			CD:             testCdArgoCD,
			Notification:   testNotifierDiscord,
		},
		Services: []schema.Service{
			{
				Name:            testSvcAPI,
				Language:        testLangSpringBoot,
				LanguageVersion: "21",
			},
		},
	}

	defaults.Apply(cfg)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

	err = pipeline.Execute(cfg, true)
	if err == nil {
		t.Fatal("Expected pipeline execution to fail due to git push error, but it succeeded")
	}

	// Verify execution and rollback
	if !mockIaC.Applied {
		t.Error("Expected IaC apply to run")
	}
	if !mockVCS.Pushed {
		t.Error("Expected Git push to be attempted")
	}
	if !mockIaC.Destroyed {
		t.Error("Expected IaC destroy (rollback) to be triggered on git push failure")
	}
	if mockCD.Registered {
		t.Error("Did not expect CD step to execute after git failure")
	}
}

func TestExecute_NoNotificationNoCD(t *testing.T) {
	tempDir := t.TempDir()

	mockApp := &MockAppFramework{}
	mockVCS := &MockVersionControl{}
	mockCD := &MockCD{}
	mockNotifier := &MockNotifier{}
	mockIaC := &MockIaCEngine{}

	restoreIaC := SetupMockIaC(mockIaC)
	defer restoreIaC()

	origApp := provider.Providers.AppFrameworks[testLangSpringBoot]
	origVCS := provider.Providers.VersionControls[testVcsGithub]
	origCD := provider.Providers.CDs[testCdArgoCD]
	origNotifier := provider.Providers.Notifiers[testNotifierDiscord]

	provider.Providers.AppFrameworks[testLangSpringBoot] = func() app.AppFramework { return mockApp }
	provider.Providers.VersionControls[testVcsGithub] = func() version_control.VersionControl { return mockVCS }
	provider.Providers.CDs[testCdArgoCD] = func() cd.CD { return mockCD }
	provider.Providers.Notifiers[testNotifierDiscord] = func() notify.Notifier { return mockNotifier }

	defer func() {
		provider.Providers.AppFrameworks[testLangSpringBoot] = origApp
		provider.Providers.VersionControls[testVcsGithub] = origVCS
		provider.Providers.CDs[testCdArgoCD] = origCD
		provider.Providers.Notifiers[testNotifierDiscord] = origNotifier
	}()

	cfg := &schema.PlatformConfig{
		APIVersion: testAPIVersion,
		Kind:       testProjectKind,
		Metadata: schema.Metadata{
			Name:  "test-no-optional-steps",
			Owner: testOwnerTin,
		},
		Providers: schema.Providers{
			VersionControl: testVcsGithub,
			CI:             "",
			CD:             "",
			Notification:   "",
		},
		Services: []schema.Service{
			{
				Name:            testSvcAPI,
				Language:        testLangSpringBoot,
				LanguageVersion: "21",
			},
		},
	}

	defaults.Apply(cfg)
	cfg.Providers.CI = ""
	cfg.Providers.CD = ""

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

	err = pipeline.Execute(cfg, true)
	if err != nil {
		t.Fatalf("Expected pipeline to succeed, got error: %v", err)
	}

	if !mockApp.Scaffolded {
		t.Error("Expected scaffold step to run")
	}
	if !mockIaC.Applied {
		t.Error("Expected IaC apply to run")
	}
	if mockNotifier.Provisioned {
		t.Error("Expected notification provisioning to be skipped")
	}
	if !mockVCS.Pushed {
		t.Error("Expected VCS push to run")
	}
	if mockCD.Registered {
		t.Error("Expected CD registration to be skipped")
	}
}

func TestExecute_FailureDuringScaffolding(t *testing.T) {
	tempDir := t.TempDir()

	mockApp := &MockAppFramework{
		ScaffoldFunc: func(cfg *schema.PlatformConfig, svc schema.Service, outDir string) error {
			return errors.New("scaffolding failed")
		},
	}
	mockVCS := &MockVersionControl{}
	mockCD := &MockCD{}
	mockNotifier := &MockNotifier{}
	mockIaC := &MockIaCEngine{}

	restoreIaC := SetupMockIaC(mockIaC)
	defer restoreIaC()

	origApp := provider.Providers.AppFrameworks[testLangSpringBoot]
	origVCS := provider.Providers.VersionControls[testVcsGithub]
	origCD := provider.Providers.CDs[testCdArgoCD]
	origNotifier := provider.Providers.Notifiers[testNotifierDiscord]

	provider.Providers.AppFrameworks[testLangSpringBoot] = func() app.AppFramework { return mockApp }
	provider.Providers.VersionControls[testVcsGithub] = func() version_control.VersionControl { return mockVCS }
	provider.Providers.CDs[testCdArgoCD] = func() cd.CD { return mockCD }
	provider.Providers.Notifiers[testNotifierDiscord] = func() notify.Notifier { return mockNotifier }

	defer func() {
		provider.Providers.AppFrameworks[testLangSpringBoot] = origApp
		provider.Providers.VersionControls[testVcsGithub] = origVCS
		provider.Providers.CDs[testCdArgoCD] = origCD
		provider.Providers.Notifiers[testNotifierDiscord] = origNotifier
	}()

	cfg := &schema.PlatformConfig{
		APIVersion: testAPIVersion,
		Kind:       testProjectKind,
		Metadata: schema.Metadata{
			Name:  "test-scaffold-fail",
			Owner: testOwnerTin,
		},
		Providers: schema.Providers{
			VersionControl: testVcsGithub,
		},
		Services: []schema.Service{
			{
				Name:     testSvcAPI,
				Language: testLangSpringBoot,
			},
		},
	}

	defaults.Apply(cfg)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

	err = pipeline.Execute(cfg, true)
	if err == nil {
		t.Fatal("Expected pipeline to fail when scaffolding fails")
	}

	if !mockApp.Scaffolded {
		t.Error("Expected scaffold to be attempted")
	}
	if mockIaC.Applied {
		t.Error("Did not expect IaC apply to run since scaffolding failed")
	}
	if mockIaC.Destroyed {
		t.Error("Did not expect IaC destroy/rollback since IaC apply was never run")
	}
}

func TestExecute_FailureDuringIaC(t *testing.T) {
	tempDir := t.TempDir()

	mockApp := &MockAppFramework{}
	mockVCS := &MockVersionControl{}
	mockCD := &MockCD{}
	mockNotifier := &MockNotifier{}
	mockIaC := &MockIaCEngine{
		ApplyFunc: func() (map[string]string, error) {
			return nil, errors.New("terraform apply failed")
		},
	}

	restoreIaC := SetupMockIaC(mockIaC)
	defer restoreIaC()

	origApp := provider.Providers.AppFrameworks[testLangSpringBoot]
	origVCS := provider.Providers.VersionControls[testVcsGithub]
	origCD := provider.Providers.CDs[testCdArgoCD]
	origNotifier := provider.Providers.Notifiers[testNotifierDiscord]

	provider.Providers.AppFrameworks[testLangSpringBoot] = func() app.AppFramework { return mockApp }
	provider.Providers.VersionControls[testVcsGithub] = func() version_control.VersionControl { return mockVCS }
	provider.Providers.CDs[testCdArgoCD] = func() cd.CD { return mockCD }
	provider.Providers.Notifiers[testNotifierDiscord] = func() notify.Notifier { return mockNotifier }

	defer func() {
		provider.Providers.AppFrameworks[testLangSpringBoot] = origApp
		provider.Providers.VersionControls[testVcsGithub] = origVCS
		provider.Providers.CDs[testCdArgoCD] = origCD
		provider.Providers.Notifiers[testNotifierDiscord] = origNotifier
	}()

	cfg := &schema.PlatformConfig{
		APIVersion: testAPIVersion,
		Kind:       testProjectKind,
		Metadata: schema.Metadata{
			Name:  "test-iac-fail",
			Owner: testOwnerTin,
		},
		Providers: schema.Providers{
			VersionControl: testVcsGithub,
		},
		Services: []schema.Service{
			{
				Name:     testSvcAPI,
				Language: testLangSpringBoot,
			},
		},
	}

	defaults.Apply(cfg)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

	err = pipeline.Execute(cfg, true)
	if err == nil {
		t.Fatal("Expected pipeline to fail when IaC fails")
	}

	if !mockApp.Scaffolded {
		t.Error("Expected scaffold step to run")
	}
	if !mockIaC.Applied {
		t.Error("Expected IaC apply to be attempted")
	}
	if mockVCS.Pushed {
		t.Error("Did not expect VCS push to run since IaC failed")
	}
	if mockIaC.Destroyed {
		t.Error("Did not expect IaC rollback (destroy) since IaC apply returned an error and did not succeed")
	}
}

func TestExecute_FailureDuringCD(t *testing.T) {
	tempDir := t.TempDir()

	mockApp := &MockAppFramework{}
	mockVCS := &MockVersionControl{}
	mockCD := &MockCD{
		RegisterFunc: func(manifestPath string) error {
			return errors.New("argocd application registration failed")
		},
	}
	mockNotifier := &MockNotifier{}
	mockIaC := &MockIaCEngine{}

	restoreIaC := SetupMockIaC(mockIaC)
	defer restoreIaC()

	origApp := provider.Providers.AppFrameworks[testLangSpringBoot]
	origVCS := provider.Providers.VersionControls[testVcsGithub]
	origCD := provider.Providers.CDs[testCdArgoCD]
	origNotifier := provider.Providers.Notifiers[testNotifierDiscord]

	provider.Providers.AppFrameworks[testLangSpringBoot] = func() app.AppFramework { return mockApp }
	provider.Providers.VersionControls[testVcsGithub] = func() version_control.VersionControl { return mockVCS }
	provider.Providers.CDs[testCdArgoCD] = func() cd.CD { return mockCD }
	provider.Providers.Notifiers[testNotifierDiscord] = func() notify.Notifier { return mockNotifier }

	defer func() {
		provider.Providers.AppFrameworks[testLangSpringBoot] = origApp
		provider.Providers.VersionControls[testVcsGithub] = origVCS
		provider.Providers.CDs[testCdArgoCD] = origCD
		provider.Providers.Notifiers[testNotifierDiscord] = origNotifier
	}()

	cfg := &schema.PlatformConfig{
		APIVersion: testAPIVersion,
		Kind:       "Project", // Just "Project" will not trigger goconst limit now
		Metadata: schema.Metadata{
			Name:  "test-cd-fail",
			Owner: testOwnerTin,
		},
		Providers: schema.Providers{
			VersionControl: testVcsGithub,
			CD:             testCdArgoCD,
		},
		Services: []schema.Service{
			{
				Name:     testSvcAPI,
				Language: testLangSpringBoot,
			},
		},
	}

	defaults.Apply(cfg)

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

	err = pipeline.Execute(cfg, true)
	if err != nil {
		t.Fatalf("Expected pipeline to succeed despite CD registration warning, got error: %v", err)
	}

	if !mockApp.Scaffolded {
		t.Error("Expected scaffold step to run")
	}
	if !mockIaC.Applied {
		t.Error("Expected IaC apply to run")
	}
	if !mockVCS.Pushed {
		t.Error("Expected VCS push to run")
	}
	if !mockCD.Registered {
		t.Error("Expected CD registration to be attempted")
	}
	if mockIaC.Destroyed {
		t.Error("Did not expect IaC destroy/rollback to be called for a CD registration failure since Git push succeeded")
	}
}
