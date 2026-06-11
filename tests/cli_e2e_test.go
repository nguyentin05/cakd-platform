package tests_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func getBinaryPath(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := filepath.Dir(wd)
	binaryPath := filepath.Join(projectRoot, "bin", "cakd")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Logf("Binary not found at %s. Compiling it first...", binaryPath)
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/cakd")
		cmd.Dir = projectRoot
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to compile binary: %v", err)
		}
	}
	return binaryPath
}

const (
	cliAPIVersion = "platform.dev/v1alpha1"
	cliProject    = "Project"
	cliOwnerTin   = "tin"
	cliGithub     = "github"
	cliSvcAPI     = "api"
	cliSpringBoot = "java-spring-boot"
)

func TestCakdCLI_ValidConfig(t *testing.T) {
	binaryPath := getBinaryPath(t)

	validYAML := "apiVersion: " + cliAPIVersion + "\n" +
		"kind: " + cliProject + "\n" +
		"metadata:\n" +
		"  name: e2e-success\n" +
		"  owner: " + cliOwnerTin + "\n" +
		"providers:\n" +
		"  versionControl: " + cliGithub + "\n" +
		"services:\n" +
		"  - name: " + cliSvcAPI + "\n" +
		"    language: " + cliSpringBoot + "\n"

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "valid_platform.yaml")
	if err := os.WriteFile(filePath, []byte(validYAML), 0600); err != nil {
		t.Fatalf("Failed to write valid yaml: %v", err)
	}

	cmd := exec.Command(binaryPath, "validate", "-f", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Expected CLI validate command to succeed, but got error: %v. Output:\n%s", err, out.String())
	}

	expectedOutput := "Valid! Project: e2e-success"
	if !strings.Contains(out.String(), expectedOutput) {
		t.Errorf("Expected output to contain %q, got: %s", expectedOutput, out.String())
	}
}

func TestCakdCLI_InvalidConfig(t *testing.T) {
	binaryPath := getBinaryPath(t)

	invalidYAML := "apiVersion: " + cliAPIVersion + "\n" +
		"kind: " + cliProject + "\n" +
		"metadata:\n" +
		"  name: e2e-failed\n" +
		"providers:\n" +
		"  versionControl: " + cliGithub + "\n" +
		"services:\n" +
		"  - name: " + cliSvcAPI + "\n" +
		"    language: " + cliSpringBoot + "\n"

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid_platform.yaml")
	if err := os.WriteFile(filePath, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("Failed to write invalid yaml: %v", err)
	}

	cmd := exec.Command(binaryPath, "validate", "-f", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI validate command to fail, but it succeeded. Output:\n%s", out.String())
	}

	expectedError := "field 'metadata.owner' is required"
	if !strings.Contains(out.String(), expectedError) {
		t.Errorf("Expected error message to contain %q, got: %s", expectedError, out.String())
	}
}

func TestCakdCLI_ValidateNonExistent(t *testing.T) {
	binaryPath := getBinaryPath(t)

	cmd := exec.Command(binaryPath, "validate", "-f", "this-file-does-not-exist.yaml")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected validate to fail with non-existent file, but it succeeded")
	}

	if !strings.Contains(out.String(), "no such file or directory") && !strings.Contains(out.String(), "Invalid:") {
		t.Errorf("Expected file not found error, got: %s", out.String())
	}
}

func TestCakdCLI_ValidateDefaultMissing(t *testing.T) {
	binaryPath := getBinaryPath(t)

	tempDir := t.TempDir()
	cmd := exec.Command(binaryPath, "validate")
	cmd.Dir = tempDir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected validate to fail in empty directory without -f, but it succeeded")
	}

	if !strings.Contains(out.String(), "no such file or directory") {
		t.Errorf("Expected file not found error, got: %s", out.String())
	}
}

func TestCakdCLI_Version(t *testing.T) {
	binaryPath := getBinaryPath(t)

	cmd := exec.Command(binaryPath, "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Expected version command to succeed, got error: %v", err)
	}

	if !strings.Contains(out.String(), "CAKD Platform CLI") {
		t.Errorf("Expected version output to contain 'CAKD Platform CLI', got: %s", out.String())
	}
}

func TestCakdCLI_InitErrorBehavior(t *testing.T) {
	binaryPath := getBinaryPath(t)

	cmd := exec.Command(binaryPath, "init", "--argocd=false", "--monitoring=false", "--logging=false")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		if !strings.Contains(out.String(), "Cluster initialization failed") {
			t.Logf("Command failed as expected but output was: %s", out.String())
		}
	} else {
		t.Log("Command completed successfully (k8s environment might be running)")
	}
}
