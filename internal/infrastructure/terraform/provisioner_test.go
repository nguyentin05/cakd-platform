package terraform_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/infrastructure/terraform"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

func TestProvisionerIsolatedStateDirectories(t *testing.T) {
	workDir := t.TempDir()
	runnerCalled := []string{}

	prov := terraform.NewProvisionerWithRunner(func(_ context.Context, dir string, args ...string) ([]byte, error) {
		runnerCalled = append(runnerCalled, dir+":"+strings.Join(args, " "))
		if len(args) > 0 && args[0] == "output" {
			return []byte(`{
				"public_ip": {"value": "54.20.30.40"},
				"private_key_pem": {"value": "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----"},
				"repo_clone_url": {"value": "https://github.com/acme/order-system.git"}
			}`), nil
		}
		return []byte("ok"), nil
	})

	ctx := context.Background()
	hostRes, err := prov.ProvisionHost(ctx, ports.CloudHostProvisionRequest{
		ProjectName:  "order-system",
		Region:       "ap-southeast-1",
		InstanceType: "t3.medium",
		DiskSize:     30,
		WorkDir:      filepath.Join(workDir, ".platform", "terraform", "cluster"),
	})
	if err != nil {
		t.Fatalf("ProvisionHost failed: %v", err)
	}
	if hostRes.PublicIP != "54.20.30.40" {
		t.Fatalf("expected PublicIP 54.20.30.40, got %s", hostRes.PublicIP)
	}
	if _, err := os.Stat(hostRes.SSHKeyPath); err != nil {
		t.Fatalf("expected SSHKeyPath %s to exist: %v", hostRes.SSHKeyPath, err)
	}

	repoURL, err := prov.ProvisionGitHubRepo(ctx, ports.RepoProvisionRequest{
		ProjectName: "order-system",
		Owner:       "acme",
		Token:       "ghp_test1234",
		StateDir:    filepath.Join(workDir, ".platform", "terraform", "vcs"),
	})
	if err != nil {
		t.Fatalf("ProvisionGitHubRepo failed: %v", err)
	}
	if repoURL != "https://github.com/acme/order-system.git" {
		t.Fatalf("unexpected repoURL: %s", repoURL)
	}

	// Verify both cluster and vcs directories have independent main.tf files
	if _, err := os.Stat(filepath.Join(workDir, ".platform", "terraform", "cluster", "main.tf")); err != nil {
		t.Fatalf("expected cluster/main.tf to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, ".platform", "terraform", "vcs", "main.tf")); err != nil {
		t.Fatalf("expected vcs/main.tf to exist: %v", err)
	}
}
