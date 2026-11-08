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

func TestValidateBidirectionalFrameworkOptions(t *testing.T) {
	svc := validate.NewService()

	// 1. Pure FrameworkOptions syncs to legacy fields
	pureOptsProj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "opt-sync", Owner: "acme"},
		Spec: project.Spec{
			Services: []project.ServiceSpec{
				{
					Name:     "payment-svc",
					Language: catalog.LangJavaSpringBoot,
					FrameworkOptions: map[string]string{
						"springBootVersion": "3.5.15.RELEASE",
						"groupId":           "com.acme.pay",
						"artifactId":        "pay-core",
						"projectBuild":      "maven-project",
						"packaging":         "jar",
						"configuration":     "yaml",
					},
				},
			},
		},
	}
	if err := svc.ValidateAndDefault(pureOptsProj); err != nil {
		t.Fatalf("ValidateAndDefault failed for pure FrameworkOptions: %v", err)
	}
	s1 := pureOptsProj.Spec.Services[0]
	if s1.SpringBootVersion != "3.5.15.RELEASE" || s1.GroupID != "com.acme.pay" || s1.ArtifactID != "pay-core" {
		t.Fatalf("legacy fields not populated from FrameworkOptions: %+v", s1)
	}
	if s1.ProjectBuild != "maven-project" || s1.Packaging != "jar" || s1.Configuration != "yaml" {
		t.Fatalf("build/packaging/config not populated from FrameworkOptions: %+v", s1)
	}

	// 2. Legacy fields sync to FrameworkOptions
	legacyProj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "legacy-sync", Owner: "acme"},
		Spec: project.Spec{
			Services: []project.ServiceSpec{
				{
					Name:              "order-svc",
					Language:          catalog.LangJavaSpringBoot,
					SpringBootVersion: "3.5.15.RELEASE",
					GroupID:           "com.acme.orders",
					ArtifactID:        "orders-core",
					ProjectBuild:      "gradle-project",
					Packaging:         "jar",
					Configuration:     "properties",
				},
			},
		},
	}
	if err := svc.ValidateAndDefault(legacyProj); err != nil {
		t.Fatalf("ValidateAndDefault failed for legacy fields: %v", err)
	}
	s2 := legacyProj.Spec.Services[0]
	if s2.FrameworkOptions["springBootVersion"] != "3.5.15.RELEASE" ||
		s2.FrameworkOptions["groupId"] != "com.acme.orders" ||
		s2.FrameworkOptions["artifactId"] != "orders-core" ||
		s2.FrameworkOptions["projectBuild"] != "gradle-project" ||
		s2.FrameworkOptions["packaging"] != "jar" ||
		s2.FrameworkOptions["configuration"] != "properties" {
		t.Fatalf("FrameworkOptions not populated from legacy fields: %+v", s2.FrameworkOptions)
	}

	// 3. Invalid FrameworkOption springBootVersion returns validation error
	invalidProj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata:   project.Metadata{Name: "invalid-opt", Owner: "acme"},
		Spec: project.Spec{
			Services: []project.ServiceSpec{
				{
					Name:     "invalid-svc",
					Language: catalog.LangJavaSpringBoot,
					FrameworkOptions: map[string]string{
						"springBootVersion": "99.99.99-INVALID",
					},
				},
			},
		},
	}
	err := svc.ValidateAndDefault(invalidProj)
	if !errcode.IsCode(err, errcode.CodeValidation) {
		t.Fatalf("expected CodeValidation for invalid springBootVersion in FrameworkOptions, got %v", err)
	}
}
