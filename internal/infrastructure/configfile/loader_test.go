package configfile_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/configfile"
)

//nolint:gocyclo
func TestLoadCanonicalAndLegacySpecs(t *testing.T) {
	loader := configfile.NewLoader()

	canonicalYAML := []byte(`
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: order-system
  owner: my-org
  environment: staging
spec:
  cluster:
    provider: aws-ec2
    region: ap-southeast-1
    domain: 1.2.3.4.nip.io
  providers:
    versionControl: github
    ci: github-actions
    cd: argocd
  services:
    - name: order-api
      language: java-spring-boot
`)
	proj, err := loader.LoadBytes(canonicalYAML)
	if err != nil {
		t.Fatalf("unexpected error loading canonical YAML: %v", err)
	}
	if proj.APIVersion != project.CanonicalAPIVersion {
		t.Fatalf("expected apiVersion %s, got %s", project.CanonicalAPIVersion, proj.APIVersion)
	}
	if proj.Spec.Cluster == nil || proj.Spec.Cluster.Provider != "aws-ec2" {
		t.Fatalf("expected spec.cluster.provider=aws-ec2, got %+v", proj.Spec.Cluster)
	}

	legacyYAML := []byte(`
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: legacy-system
  owner: legacy-org
infra:
  provider: aws-ec2
  host: 54.1.2.3
providers:
  versionControl: github
  ci: github-actions
  cd: argocd
backing:
  - name: main-db
    type: postgresql
services:
  - name: legacy-api
    language: java-spring-boot
    uses: [main-db]
`)
	legacyProj, err := loader.LoadBytes(legacyYAML)
	if err != nil {
		t.Fatalf("unexpected error loading legacy YAML: %v", err)
	}
	if legacyProj.Spec.Cluster == nil || legacyProj.Spec.Cluster.Host != "54.1.2.3" {
		t.Fatalf("expected legacy infra mapped to spec.cluster, got %+v", legacyProj.Spec.Cluster)
	}
	if len(legacyProj.Spec.Backing) != 1 || legacyProj.Spec.Backing[0].Name != "main-db" {
		t.Fatalf("expected legacy backing mapped to spec.backing, got %+v", legacyProj.Spec.Backing)
	}
	if len(legacyProj.Spec.Services) != 1 || legacyProj.Spec.Services[0].Name != "legacy-api" {
		t.Fatalf("expected legacy services mapped to spec.services, got %+v", legacyProj.Spec.Services)
	}

	outPath := filepath.Join(t.TempDir(), "platform.yaml")
	if err := loader.SaveFile(outPath, legacyProj); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}
	reloaded, err := loader.LoadFile(outPath)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if reloaded.APIVersion != project.CanonicalAPIVersion {
		t.Fatalf("expected saved file to normalize to canonical apiVersion, got %s", reloaded.APIVersion)
	}
	data, err := loader.MarshalCanonical(reloaded)
	if err != nil {
		t.Fatalf("MarshalCanonical failed: %v", err)
	}
	if !strings.Contains(string(data), "spec:") {
		t.Fatalf("expected canonical YAML output to contain 'spec:' block, got:\n%s", string(data))
	}
}

func TestLoadGenericFrameworkOptionsAndLegacyFallback(t *testing.T) {
	loader := configfile.NewLoader()

	yamlContent := []byte(`
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: polyglot-platform
  owner: acme
spec:
  services:
    - name: generic-service
      language: node-express
      frameworkOptions:
        nodeVersion: "20"
        packageManager: pnpm
        framework: express
    - name: legacy-spring-service
      language: java-spring-boot
      springBootVersion: "3.2.0"
      groupId: com.acme
      artifactId: order-service
`)
	proj, err := loader.LoadBytes(yamlContent)
	if err != nil {
		t.Fatalf("LoadBytes failed: %v", err)
	}
	if len(proj.Spec.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(proj.Spec.Services))
	}

	// 1. Generic service with FrameworkOptions
	s1 := proj.Spec.Services[0]
	if got := s1.GetFrameworkOption("nodeVersion", ""); got != "20" {
		t.Fatalf("expected nodeVersion=20, got %q", got)
	}
	if got := s1.GetFrameworkOption("packageManager", ""); got != "pnpm" {
		t.Fatalf("expected packageManager=pnpm, got %q", got)
	}
	if got := s1.GetFrameworkOption("missing", "default-val"); got != "default-val" {
		t.Fatalf("expected fallback 'default-val', got %q", got)
	}

	// 2. Legacy Spring Boot service
	s2 := proj.Spec.Services[1]
	if got := s2.GetFrameworkOption("springBootVersion", ""); got != "3.2.0" {
		t.Fatalf("expected springBootVersion=3.2.0, got %q", got)
	}
	if got := s2.GetFrameworkOption("groupId", ""); got != "com.acme" {
		t.Fatalf("expected groupId=com.acme, got %q", got)
	}
	if got := s2.GetFrameworkOption("artifactId", ""); got != "order-service" {
		t.Fatalf("expected artifactId=order-service, got %q", got)
	}
}
