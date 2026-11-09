package createproject_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
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
	pushed      bool
	passedToken string
}

func (f *fakeVCS) ProvisionAndPush(_ context.Context, req ports.RepoProvisionRequest) (string, error) {
	f.pushed = true
	f.passedToken = req.Token
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
				{Name: "order-api", Language: catalog.LangJavaSpringBoot},
			},
		},
	}

	outDir := filepath.Join(tmp, "out", "order-system")
	res, err := uc.Execute(context.Background(), createproject.Request{
		Project:      proj,
		OutputDir:    outDir,
		Force:        true,
		ScaffoldOnly: false,
		VCSToken:     "vcs_custom_token",
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
	if vcs.passedToken != "vcs_custom_token" {
		t.Fatalf("expected VCSToken to be passed, got %q", vcs.passedToken)
	}
	if _, err := os.Stat(filepath.Join(outDir, "services/order-api/helm/values.yaml")); err != nil {
		t.Fatalf("expected rendered helm values.yaml: %v", err)
	}
}

func TestCreateProjectLocalGitWithoutToken(t *testing.T) {
	tmp := t.TempDir()
	vcs := &fakeVCS{}
	// Empty credStore with no tokens
	credStore := credstore.NewStoreWithPath(filepath.Join(tmp, "creds.yaml"))

	uc := createproject.NewService(
		configfile.NewLoader(),
		validate.NewService(),
		fakeFrameworkGen{},
		scaffold.NewRenderer(),
		vcs,
		nil,
		nil,
		credStore,
		nil,
	)

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "local-proj", Owner: "localdev"},
		Spec: project.Spec{
			Providers: &project.ProvidersSpec{
				VersionControl: catalog.ProviderLocalGit,
			},
			Services: []project.ServiceSpec{
				{Name: "svc-a", Language: catalog.LangJavaSpringBoot},
			},
		},
	}

	outDir := filepath.Join(tmp, "out", "local-proj")
	res, err := uc.Execute(context.Background(), createproject.Request{
		Project:   proj,
		OutputDir: outDir,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Execute should succeed for local-git without token, got err: %v", err)
	}
	if !vcs.pushed {
		t.Fatalf("expected vcs ProvisionAndPush to be called for local-git")
	}
	if vcs.passedToken != "" {
		t.Fatalf("expected empty token for local-git, got %q", vcs.passedToken)
	}
	if res.ProjectName != "local-proj" {
		t.Fatalf("unexpected project name: %s", res.ProjectName)
	}
}

func TestCreateProjectMissingTokenForRemoteVCSErrors(t *testing.T) {
	tmp := t.TempDir()
	vcs := &fakeVCS{}
	// Empty credStore
	credStore := credstore.NewStoreWithPath(filepath.Join(tmp, "creds.yaml"))

	uc := createproject.NewService(
		configfile.NewLoader(),
		validate.NewService(),
		fakeFrameworkGen{},
		scaffold.NewRenderer(),
		vcs,
		nil,
		nil,
		credStore,
		nil,
	)

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "remote-proj", Owner: "myorg"},
		Spec: project.Spec{
			Providers: &project.ProvidersSpec{
				VersionControl: catalog.ProviderGitHub,
			},
			Services: []project.ServiceSpec{
				{Name: "svc-a", Language: catalog.LangJavaSpringBoot},
			},
		},
	}

	outDir := filepath.Join(tmp, "out", "remote-proj")
	_, err := uc.Execute(context.Background(), createproject.Request{
		Project:   proj,
		OutputDir: outDir,
		Force:     true,
	})
	if err == nil {
		t.Fatalf("expected error when token is missing for remote VCS")
	}
	if !strings.Contains(err.Error(), "github token is required to provision and push the repository") {
		t.Fatalf("expected error message to mention github token, got %v", err)
	}
}
