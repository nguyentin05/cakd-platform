package createproject_test

import (
	"context"
	"os"
	"path/filepath"
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

type fakeFrameworkGen struct{}

func (fakeFrameworkGen) GenerateService(_ context.Context, _ project.ServiceSpec, destDir string) error {
	return os.WriteFile(filepath.Join(destDir, "pom.xml"), []byte("<project/>"), 0o600)
}

type fakeVCS struct {
	pushed bool
}

func (f *fakeVCS) ProvisionAndPush(_ context.Context, _ ports.RepoProvisionRequest) (string, error) {
	f.pushed = true
	return "https://github.com/acme/order-system.git", nil
}

func (f *fakeVCS) Rollback(_ context.Context, _ ports.RepoProvisionRequest) error {
	return nil
}

func TestCreateProjectScaffoldAndFullPipeline(t *testing.T) {
	tmp := t.TempDir()
	stateStore := statestore.NewStoreWithDir(tmp)
	_ = stateStore.SaveClusterState(ports.ClusterState{
		Provider: "aws-ec2",
		Host:     "54.99.88.77",
		Domain:   "54.99.88.77.nip.io",
	})
	credStore := credstore.NewStoreWithPath(filepath.Join(tmp, "creds.yaml"))
	_ = credStore.Save(ports.CredentialSet{GitHubToken: "ghp_testtoken12345"})

	vcs := &fakeVCS{}
	uc := createproject.NewService(
		configfile.NewLoader(),
		validate.NewService(),
		fakeFrameworkGen{},
		scaffold.NewRenderer(),
		vcs,
		nil,
		nil,
		credStore,
		stateStore,
	)

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "order-system", Owner: "acme"},
		Spec: project.Spec{
			Services: []project.ServiceSpec{
				{Name: "order-api", Language: "java-spring-boot"},
			},
		},
	}

	outDir := filepath.Join(tmp, "out", "order-system")
	res, err := uc.Execute(context.Background(), createproject.Request{
		Project:      proj,
		OutputDir:    outDir,
		Force:        true,
		ScaffoldOnly: false,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if res.Domain != "54.99.88.77.nip.io" {
		t.Fatalf("expected domain from ClusterStateStore 54.99.88.77.nip.io, got %s", res.Domain)
	}
	if !vcs.pushed {
		t.Fatalf("expected VCS ProvisionAndPush to be called")
	}
	if _, err := os.Stat(filepath.Join(outDir, "services/order-api/helm/values.yaml")); err != nil {
		t.Fatalf("expected rendered helm values.yaml: %v", err)
	}
}
