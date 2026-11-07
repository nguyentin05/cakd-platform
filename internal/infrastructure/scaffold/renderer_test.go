package scaffold_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/scaffold"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

func TestRendererGeneratesProjectAndServiceTemplates(t *testing.T) {
	outDir := t.TempDir()
	renderer := scaffold.NewRenderer()

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata: project.Metadata{
			Name:  "order-system",
			Owner: "acme",
		},
		Spec: project.Spec{
			Providers: &project.ProvidersSpec{
				VersionControl: "github",
				CI:             "github-actions",
				CD:             "argocd",
			},
			Backing: []project.BackingSpec{
				{Name: "orders-db", Type: "postgresql", Version: "16", Storage: "5Gi"},
			},
			Services: []project.ServiceSpec{
				{
					Name:          "order-api",
					Language:      "java-spring-boot",
					Version:       "21",
					ProjectBuild:  "maven-project",
					Configuration: "yaml",
					Replicas:      2,
					Resources:     &project.ResourcesSpec{CPU: "500m", Memory: "512Mi"},
					Uses:          []string{"orders-db"},
				},
			},
		},
	}

	err := renderer.RenderProject(context.Background(), ports.ScaffoldRequest{
		Project:    proj,
		BaseDomain: "54.1.2.3.nip.io",
		OutputDir:  outDir,
	})
	if err != nil {
		t.Fatalf("RenderProject failed: %v", err)
	}

	expectedFiles := []string{
		".gitignore",
		".github/workflows/ci.yml",
		"deploy/application.yaml",
		"services/order-api/Dockerfile",
		"services/order-api/helm/Chart.yaml",
		"services/order-api/helm/values.yaml",
		"services/order-api/helm/templates/deployment.yaml",
		"services/order-api/helm/templates/service.yaml",
		"services/order-api/helm/templates/ingress.yaml",
		"services/order-api/src/main/resources/application.yml",
		"services/order-api/src/test/resources/application.yml",
	}
	for _, rel := range expectedFiles {
		p := filepath.Join(outDir, rel)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected generated file %s to exist: %v", rel, err)
		}
	}

	valuesBytes, err := os.ReadFile(filepath.Join(outDir, "services/order-api/helm/values.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(valuesBytes), "54.1.2.3.nip.io") {
		t.Fatalf("expected values.yaml to contain domain 54.1.2.3.nip.io, got:\n%s", string(valuesBytes))
	}
}
