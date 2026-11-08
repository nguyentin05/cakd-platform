package github_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/github"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

type mockProvisioner struct {
	provisionURL string
	provisionErr error
	destroyed    bool
	destroyErr   error
}

func (m *mockProvisioner) ProvisionGitHubRepo(_ context.Context, _ ports.RepoProvisionRequest) (string, error) {
	if m.provisionErr != nil {
		return "", m.provisionErr
	}
	return m.provisionURL, nil
}

func (m *mockProvisioner) DestroyGitHubRepo(_ context.Context, _ ports.RepoProvisionRequest) error {
	m.destroyed = true
	return m.destroyErr
}

func TestClient_ProvisionAndRollback(t *testing.T) {
	mock := &mockProvisioner{
		provisionURL: "https://github.com/acme/order-api.git",
	}
	client := github.NewClient(mock)

	// Step 1: Rollback invokes DestroyGitHubRepo
	err := client.Rollback(context.Background(), ports.RepoProvisionRequest{
		ProjectName: "order-api",
	})
	if err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}
	if !mock.destroyed {
		t.Fatal("expected DestroyGitHubRepo to have been called")
	}

	// Step 2: Unconfigured provisioner returns error
	nilClient := github.NewClient(nil)
	_, err = nilClient.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{})
	if err == nil {
		t.Fatal("expected error with nil provisioner, got nil")
	}

	// Step 3: Provision error returns immediately
	errProv := &mockProvisioner{provisionErr: errors.New("terraform quota exceeded")}
	errClient := github.NewClient(errProv)
	_, err = errClient.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{})
	if err == nil || err.Error() != "terraform quota exceeded" {
		t.Fatalf("expected quota error, got: %v", err)
	}
}

func TestClient_PathTraversalRejection(t *testing.T) {
	mock := &mockProvisioner{
		provisionURL: "https://github.com/acme/order-api.git",
	}
	client := github.NewClient(mock)

	// Directory path with .. must be rejected
	_, err := client.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{
		SourceDir: "../escape/dir",
		Token:     "ghp_test123",
	})
	if err == nil {
		t.Fatal("expected error for directory traversal, got nil")
	}
}

func TestClient_SuccessWithLocalGit(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "my-service")
	if err := os.MkdirAll(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "README.md"), []byte("# Test"), 0o600); err != nil {
		t.Fatal(err)
	}

	mock := &mockProvisioner{
		provisionURL: "https://github.com/acme/my-service.git",
	}
	client := github.NewClient(mock)

	// In test environments without git remote, push will fail and trigger rollback
	_, err := client.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{
		SourceDir: sourceDir,
		Token:     "dummy-token",
	})
	if err == nil {
		t.Fatal("expected push to non-existent remote to fail")
	}
	if !mock.destroyed {
		t.Fatal("expected rollback to have been triggered on git push failure")
	}
}
