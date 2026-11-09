package localgit_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/localgit"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

func TestLocalGitProvisionAndPush(t *testing.T) {
	tmp := t.TempDir()
	testFile := filepath.Join(tmp, "platform.yaml")
	_ = os.WriteFile(testFile, []byte("apiVersion: cakd.io/v1alpha1\n"), 0o600)

	client := localgit.NewClient()
	repoURL, err := client.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{
		ProjectName: "test-proj",
		Owner:       "localdev",
		SourceDir:   tmp,
	})
	if err != nil {
		t.Fatalf("ProvisionAndPush failed: %v", err)
	}

	if !strings.HasPrefix(repoURL, "file://") {
		t.Fatalf("expected file:// url, got %q", repoURL)
	}

	gitDir := filepath.Join(tmp, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Fatalf("expected .git directory to exist: %v", err)
	}

	// Test rollback removes .git
	if err := client.Rollback(context.Background(), ports.RepoProvisionRequest{
		SourceDir: tmp,
	}); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
		t.Fatalf("expected .git directory to be removed after rollback")
	}
}

func TestLocalGitPathTraversal(t *testing.T) {
	client := localgit.NewClient()
	_, err := client.ProvisionAndPush(context.Background(), ports.RepoProvisionRequest{
		SourceDir: "../secret/path",
	})
	if err == nil {
		t.Fatal("expected error on path traversal")
	}
}
