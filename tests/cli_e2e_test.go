package tests_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce  sync.Once
	builtBin   string
	buildError error
)

func getBinaryPath(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			buildError = err
			return
		}
		projectRoot := filepath.Dir(wd)
		builtBin = filepath.Join(projectRoot, "bin", "cakd-cli")
		cmd := exec.Command("go", "build", "-o", builtBin, "./cmd/cakd-cli")
		cmd.Dir = projectRoot
		buildError = cmd.Run()
	})
	if buildError != nil {
		t.Fatalf("failed to compile cakd-cli binary: %v", buildError)
	}
	return builtBin
}

func TestCakdCLI_ValidCanonicalAndLegacyConfigs(t *testing.T) {
	bin := getBinaryPath(t)
	tmp := t.TempDir()

	canonicalYAML := `apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: e2e-canonical
  owner: acme
spec:
  cluster:
    provider: local
  services:
    - name: api
      language: java-spring-boot
`
	canonicalFile := filepath.Join(tmp, "canonical.yaml")
	if err := os.WriteFile(canonicalFile, []byte(canonicalYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "validate", "-f", canonicalFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected canonical YAML validation to succeed, got %v:\n%s", err, string(out))
	}
	if !strings.Contains(string(out), "e2e-canonical") {
		t.Fatalf("unexpected output: %s", string(out))
	}

	legacyYAML := `apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: e2e-legacy
  owner: acme
providers:
  versionControl: github
services:
  - name: api
    language: java-spring-boot
`
	legacyFile := filepath.Join(tmp, "legacy.yaml")
	if err := os.WriteFile(legacyFile, []byte(legacyYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd2 := exec.Command(bin, "validate", "-f", legacyFile, "--json")
	out2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("expected legacy YAML validation to succeed, got %v:\n%s", err, string(out2))
	}
	if !strings.Contains(string(out2), `"valid": true`) {
		t.Fatalf("expected JSON valid: true output, got: %s", string(out2))
	}
}

func TestCakdCLI_InvalidConfigAndVersion(t *testing.T) {
	bin := getBinaryPath(t)
	tmp := t.TempDir()

	invalidYAML := `apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: missing-owner
spec:
  services:
    - name: api
      language: java-spring-boot
`
	invalidFile := filepath.Join(tmp, "invalid.yaml")
	if err := os.WriteFile(invalidFile, []byte(invalidYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "validate", "-f", invalidFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected validation to fail for missing metadata.owner")
	}
	if !strings.Contains(out.String(), "metadata.owner") {
		t.Fatalf("expected error mentioning metadata.owner, got: %s", out.String())
	}

	verCmd := exec.Command(bin, "version")
	verOut, err := verCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(string(verOut), "cakd-cli version") {
		t.Fatalf("unexpected version output: %s", string(verOut))
	}
}
