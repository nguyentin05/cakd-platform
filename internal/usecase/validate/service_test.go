package validate_test

import (
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

const (
	testOwner          = "acme"
	defaultBootVersion = "3.5.15.RELEASE"
	defaultPackaging   = "jar"
)

func TestValidateAppliesDefaultsAndChecksRules(t *testing.T) {
	svc := validate.NewService()

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata: project.Metadata{
			Name:  "order-system",
			Owner: testOwner,
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
		Metadata:   project.Metadata{Name: "cycled", Owner: testOwner},
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

//nolint:gocyclo
func TestValidateBidirectionalFrameworkOptions(t *testing.T) {
	svc := validate.NewService()

	t.Run("PureFrameworkOptions", func(t *testing.T) {
		proj := &project.Project{
			APIVersion: project.CanonicalAPIVersion,
			Kind:       project.KindProject,
			Metadata:   project.Metadata{Name: "opt-sync", Owner: testOwner},
			Spec: project.Spec{
				Services: []project.ServiceSpec{
					{
						Name:     "payment-svc",
						Language: catalog.LangJavaSpringBoot,
						FrameworkOptions: map[string]string{
							"springBootVersion": defaultBootVersion,
							"groupId":           "com.acme.pay",
							"artifactId":        "pay-core",
							"projectBuild":      "maven-project",
							"packaging":         defaultPackaging,
							"configuration":     "yaml",
						},
					},
				},
			},
		}
		if err := svc.ValidateAndDefault(proj); err != nil {
			t.Fatalf("ValidateAndDefault failed for pure FrameworkOptions: %v", err)
		}
		s := proj.Spec.Services[0]
		if s.SpringBootVersion != defaultBootVersion || s.GroupID != "com.acme.pay" || s.ArtifactID != "pay-core" {
			t.Fatalf("legacy fields not populated from FrameworkOptions: %+v", s)
		}
		if s.ProjectBuild != "maven-project" || s.Packaging != defaultPackaging || s.Configuration != "yaml" {
			t.Fatalf("build/packaging/config not populated from FrameworkOptions: %+v", s)
		}
	})

	t.Run("LegacyFields", func(t *testing.T) {
		proj := &project.Project{
			APIVersion: project.CanonicalAPIVersion,
			Kind:       project.KindProject,
			Metadata:   project.Metadata{Name: "legacy-sync", Owner: testOwner},
			Spec: project.Spec{
				Services: []project.ServiceSpec{
					{
						Name:              "order-svc",
						Language:          catalog.LangJavaSpringBoot,
						SpringBootVersion: defaultBootVersion,
						GroupID:           "com.acme.orders",
						ArtifactID:        "orders-core",
						ProjectBuild:      "gradle-project",
						Packaging:         defaultPackaging,
						Configuration:     "properties",
					},
				},
			},
		}
		if err := svc.ValidateAndDefault(proj); err != nil {
			t.Fatalf("ValidateAndDefault failed for legacy fields: %v", err)
		}
		s := proj.Spec.Services[0]
		if s.FrameworkOptions["springBootVersion"] != defaultBootVersion ||
			s.FrameworkOptions["groupId"] != "com.acme.orders" ||
			s.FrameworkOptions["artifactId"] != "orders-core" ||
			s.FrameworkOptions["projectBuild"] != "gradle-project" ||
			s.FrameworkOptions["packaging"] != defaultPackaging ||
			s.FrameworkOptions["configuration"] != "properties" {
			t.Fatalf("FrameworkOptions not populated from legacy fields: %+v", s.FrameworkOptions)
		}
	})

	t.Run("InvalidFrameworkOptions", func(t *testing.T) {
		proj := &project.Project{
			APIVersion: project.CanonicalAPIVersion,
			Kind:       project.KindProject,
			Metadata:   project.Metadata{Name: "invalid-opt", Owner: testOwner},
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
		err := svc.ValidateAndDefault(proj)
		if !errcode.IsCode(err, errcode.CodeValidation) {
			t.Fatalf("expected CodeValidation for invalid springBootVersion in FrameworkOptions, got %v", err)
		}
	})
}
