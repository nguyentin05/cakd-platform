package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/platform/redact"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// RepoProvisioner provisions and destroys remote Git repositories via infrastructure tools.
type RepoProvisioner interface {
	ProvisionGitHubRepo(ctx context.Context, req ports.RepoProvisionRequest) (string, error)
	DestroyGitHubRepo(ctx context.Context, req ports.RepoProvisionRequest) error
}

// Client implements ports.VCSProvider for GitHub using isolated Terraform state and redacted git CLI calls.
type Client struct {
	prov RepoProvisioner
}

// NewClient creates a GitHub VCS provider Client.
func NewClient(prov RepoProvisioner) *Client {
	return &Client{prov: prov}
}

// ProvisionAndPush provisions the GitHub repository via Terraform and pushes the scaffolded project directory.
func (c *Client) ProvisionAndPush(ctx context.Context, req ports.RepoProvisionRequest) (string, error) {
	progress.Emit(ctx, "vcs-terraform", progress.StatusRunning, "provisioning GitHub repository via Terraform")
	if c.prov == nil {
		return "", errcode.New(errcode.CodeProvision, "repository provisioner is not configured")
	}
	repoURL, err := c.prov.ProvisionGitHubRepo(ctx, req)
	if err != nil {
		return "", err
	}

	progress.Emit(ctx, "vcs-git", progress.StatusRunning, "initializing local git repository and pushing to "+repoURL)
	if err := c.initAndPushGit(ctx, req.SourceDir, repoURL, req.Token); err != nil {
		progress.Emit(ctx, "vcs-rollback", progress.StatusWarn, "git push failed; rolling back provisioned GitHub repository")
		_ = c.Rollback(ctx, req)
		return "", err
	}

	progress.Emit(ctx, "vcs-git", progress.StatusDone, "pushed scaffolded project to "+repoURL)
	return repoURL, nil
}

// Rollback destroys the provisioned GitHub repository in the isolated VCS Terraform state directory.
func (c *Client) Rollback(ctx context.Context, req ports.RepoProvisionRequest) error {
	if c.prov == nil {
		return nil
	}
	return c.prov.DestroyGitHubRepo(ctx, req)
}

func (c *Client) initAndPushGit(ctx context.Context, projectDir, repoCloneURL, token string) error {
	cleanDir := filepath.Clean(projectDir)
	if strings.Contains(cleanDir, "..") {
		return errcode.New(errcode.CodeValidation, "invalid project directory path")
	}
	gitDir := filepath.Join(cleanDir, ".git")
	if strings.Contains(gitDir, "..") || !strings.HasPrefix(gitDir, cleanDir+string(filepath.Separator)) {
		return errcode.New(errcode.CodeValidation, "invalid git directory path")
	}
	gitDir = strings.ReplaceAll(gitDir, "..", "")
	_ = os.RemoveAll(gitDir)

	steps := [][]string{
		{"init"},
		{"add", "."},
		{"commit", "-m", "chore: initial bootstrap by CAKD Platform CLI"},
		{"branch", "-M", "main"},
	}

	for _, args := range steps {
		if err := runGitRedacted(ctx, cleanDir, token, args...); err != nil {
			return err
		}
	}

	authedURL := insertTokenInURL(repoCloneURL, token)
	if err := runGitRedacted(ctx, cleanDir, token, "remote", "add", "origin", authedURL); err != nil {
		return err
	}
	defer func() {
		_ = runGitRedacted(ctx, cleanDir, token, "remote", "set-url", "origin", repoCloneURL)
	}()

	return runGitRedacted(ctx, cleanDir, token, "push", "-f", "-u", "origin", "main")
}

func insertTokenInURL(cloneURL, token string) string {
	const httpsPrefix = "https://"
	if token != "" && strings.HasPrefix(cloneURL, httpsPrefix) {
		return httpsPrefix + token + "@" + strings.TrimPrefix(cloneURL, httpsPrefix)
	}
	return cloneURL
}

func runGitRedacted(ctx context.Context, dir, token string, args ...string) error {
	//nolint:gosec
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		sanitized := redact.Sanitize(string(out), token)
		return errcode.New(errcode.CodeExternal, fmt.Sprintf("git %s failed: %s", args[0], strings.TrimSpace(sanitized)))
	}
	return nil
}
