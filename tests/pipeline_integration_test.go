package tests_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/configfile"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/credstore"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/scaffold"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/statestore"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

type mockFrameworkGen struct{}

func (mockFrameworkGen) GenerateService(_ context.Context, _ project.ServiceSpec, destDir string) error {
	return os.WriteFile(filepath.Join(destDir, "pom.xml"), []byte("<project/>"), 0o600)
}

type mockVCS struct {
	failPush     bool
	rolledBack   bool
	provisionDir string
}

func (m *mockVCS) ProvisionAndPush(ctx context.Context, req ports.RepoProvisionRequest) (string, error) {
	m.provisionDir = req.StateDir
	if m.failPush {
		_ = m.Rollback(ctx, req)
		return "", os.ErrPermission
	}
	return "https://github.com/acme/e2e-project.git", nil
}

func (m *mockVCS) Rollback(_ context.Context, _ ports.RepoProvisionRequest) error {
	m.rolledBack = true
	return nil
}

type mockGitOps struct {
	registered bool
}

func (m *mockGitOps) RegisterApplication(_ context.Context, _ ports.GitOpsRegisterRequest) error {
	m.registered = true
	return nil
}

func TestEndToEndCreateProjectAndRollbackIsolation(t *testing.T) {
	tmp := t.TempDir()
	specRepo := configfile.NewLoader()
	valUC := validate.NewService()
	renderer := scaffold.NewRenderer()
	creds := credstore.NewStoreWithPath(filepath.Join(tmp, "credentials.yaml"))
	_ = creds.Save(ports.CredentialSet{GitHubToken: "ghp_e2etoken123456"})
	states := statestore.NewStoreWithDir(tmp)
	_ = states.SaveClusterState(ports.ClusterState{
		Provider: "aws-ec2",
		Host:     "54.11.22.33",
		Domain:   "54.11.22.33.nip.io",
	})

	vcs := &mockVCS{}
	gitops := &mockGitOps{}
	createUC := createproject.NewService(specRepo, valUC, mockFrameworkGen{}, renderer, vcs, gitops, nil, creds, states)

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "e2e-project", Owner: "acme"},
		Spec: project.Spec{
			Backing: []project.BackingSpec{
				{Name: "db", Type: "postgresql"},
			},
			Services: []project.ServiceSpec{
				{Name: "billing-api", Language: "java-spring-boot", Uses: []string{"db"}},
			},
		},
	}

	outDir := filepath.Join(tmp, "out", "e2e-project")
	res, err := createUC.Execute(context.Background(), createproject.Request{
		Project:   proj,
		OutputDir: outDir,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !gitops.registered {
		t.Fatalf("expected GitOps application to be registered")
	}
	if !strings.HasSuffix(filepath.ToSlash(vcs.provisionDir), ".platform/terraform/vcs") {
		t.Fatalf("expected isolated VCS terraform state dir, got %s", vcs.provisionDir)
	}
	if res.Domain != "54.11.22.33.nip.io" {
		t.Fatalf("expected domain 54.11.22.33.nip.io, got %s", res.Domain)
	}
}
