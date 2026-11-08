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

func TestPipeline_PureFrameworkOptionsManifest(t *testing.T) {
	tmp := t.TempDir()
	specRepo := configfile.NewLoader()
	valUC := validate.NewService()
	renderer := scaffold.NewRenderer()
	creds := credstore.NewStoreWithPath(filepath.Join(tmp, "credentials.yaml"))
	_ = creds.Save(ports.CredentialSet{GitHubToken: "ghp_test123456"})
	states := statestore.NewStoreWithDir(tmp)
	_ = states.SaveClusterState(ports.ClusterState{
		Provider: "local",
		Domain:   "local.nip.io",
	})

	vcs := &mockVCS{}
	gitops := &mockGitOps{}
	createUC := createproject.NewService(specRepo, valUC, mockFrameworkGen{}, renderer, vcs, gitops, nil, creds, states)

	pureYAML := `apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: pure-framework-system
  owner: acme-corp
spec:
  cluster:
    provider: local
  providers:
    versionControl: github
    ci: github-actions
    cd: argocd
  services:
    - name: billing-service
      language: java-spring-boot
      frameworkOptions:
        springBootVersion: "3.5.15.RELEASE"
        groupId: "com.acme.billing"
        artifactId: "billing-core"
        projectBuild: "maven-project"
        packaging: "jar"
        configuration: "yaml"
`
	manifestPath := filepath.Join(tmp, "platform.yaml")
	if err := os.WriteFile(manifestPath, []byte(pureYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	// 1. Loading & Parsing
	proj, err := specRepo.LoadFile(manifestPath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	// 2. Validation & Bidirectional Sync PASS
	if err := valUC.ValidateAndDefault(proj); err != nil {
		t.Fatalf("ValidateAndDefault failed for pure frameworkOptions: %v", err)
	}
	svc := proj.Spec.Services[0]
	if svc.SpringBootVersion != "3.5.15.RELEASE" || svc.GroupID != "com.acme.billing" || svc.ArtifactID != "billing-core" {
		t.Fatalf("legacy fields not populated from frameworkOptions: %+v", svc)
	}

	// 3. Scaffolding PASS & Correct Files Generated
	outDir := filepath.Join(tmp, "out", "pure-framework-system")
	res, err := createUC.Execute(context.Background(), createproject.Request{
		Project:   proj,
		OutputDir: outDir,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("createUC.Execute failed: %v", err)
	}
	if res.ProjectName != "pure-framework-system" {
		t.Fatalf("expected project name pure-framework-system, got %s", res.ProjectName)
	}

	// Verify generated pom.xml and configuration
	pomPath := filepath.Join(outDir, "services", "billing-service", "pom.xml")
	if _, err := os.Stat(pomPath); err != nil {
		t.Fatalf("expected scaffolded pom.xml at %s: %v", pomPath, err)
	}

	appYamlPath := filepath.Join(outDir, "services", "billing-service", "src", "main", "resources", "application.yml")
	if _, err := os.Stat(appYamlPath); err != nil {
		t.Fatalf("expected scaffolded application.yml at %s: %v", appYamlPath, err)
	}

	// 4. Schema verification: ensure schemas/platform.schema.json defines frameworkOptions
	schemaBytes, err := os.ReadFile(filepath.Join("..", "schemas", "platform.schema.json"))
	if err != nil {
		// When test runs from repo root
		schemaBytes, err = os.ReadFile(filepath.Join("schemas", "platform.schema.json"))
	}
	if err == nil {
		if !strings.Contains(string(schemaBytes), `"frameworkOptions"`) {
			t.Fatalf("expected platform.schema.json to contain frameworkOptions definition")
		}
	}
}
