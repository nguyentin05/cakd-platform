package validate_test

import (
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

func TestValidateAppliesDefaultsAndChecksRules(t *testing.T) {
	svc := validate.NewService()

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata: project.Metadata{
			Name:  "order-system",
			Owner: "acme",
		},
		Spec: project.Spec{
			Backing: []project.BackingSpec{
				{Name: "orders-db", Type: "postgresql"},
			},
			Services: []project.ServiceSpec{
				{
					Name:     "order-api",
					Language: catalog.LangJavaSpringBoot,
					Uses:     []string{"orders-db"},
				},
			},
		},
	}

	if err := svc.ValidateAndDefault(proj); err != nil {
		t.Fatalf("expected valid project, got error: %v", err)
	}
	if proj.Spec.Cluster == nil || proj.Spec.Cluster.Provider != "local" {
		t.Fatalf("expected default cluster provider 'local', got %+v", proj.Spec.Cluster)
	}
	if proj.Spec.Backing[0].Version != "16" || proj.Spec.Backing[0].Storage != "5Gi" {
		t.Fatalf("expected default postgres 16 and 5Gi, got %+v", proj.Spec.Backing[0])
	}
	if proj.Spec.Services[0].Version != "21" || proj.Spec.Services[0].Replicas != 1 {
		t.Fatalf("expected default service version 21 and replicas 1, got %+v", proj.Spec.Services[0])
	}

	// Unknown backing reference should fail with CodeValidation
	badProj := *proj
	badProj.Spec.Services = []project.ServiceSpec{
		{Name: "order-api", Language: catalog.LangJavaSpringBoot, Uses: []string{"non-existent"}},
	}
	if err := svc.ValidateAndDefault(&badProj); !errcode.IsCode(err, errcode.CodeValidation) {
		t.Fatalf("expected CodeValidation for unknown backing ref, got %v", err)
	}
}

func TestValidateDetectsServiceDependencyCycle(t *testing.T) {
	svc := validate.NewService()
	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "cycled", Owner: "acme"},
		Spec: project.Spec{
			Services: []project.ServiceSpec{
				{Name: "svc-a", Language: catalog.LangJavaSpringBoot, Uses: []string{"svc-b"}},
				{Name: "svc-b", Language: catalog.LangJavaSpringBoot, Uses: []string{"svc-a"}},
			},
		},
	}
	if err := svc.ValidateAndDefault(proj); !errcode.IsCode(err, errcode.CodeValidation) {
		t.Fatalf("expected cycle detection CodeValidation error, got %v", err)
	}
}
