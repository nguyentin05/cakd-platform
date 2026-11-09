package localgit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// Client implements ports.VCSProvider for local git repositories without remote push.
type Client struct{}

// NewClient creates a new local git VCS provider client.
func NewClient() *Client {
	return &Client{}
}

var _ ports.VCSProvider = (*Client)(nil)

// ProvisionAndPush initializes a local git repository in req.SourceDir, commits scaffolded files, and returns a file:// URL.
func (c *Client) ProvisionAndPush(ctx context.Context, req ports.RepoProvisionRequest) (string, error) {
	progress.Emit(ctx, "vcs-local-git", progress.StatusRunning, "initializing local git repository in "+req.SourceDir)

	cleanDir := filepath.Clean(req.SourceDir)
	if strings.Contains(cleanDir, "..") {
		return "", errcode.New(errcode.CodeValidation, "invalid source directory path")
	}

	gitDir := filepath.Join(cleanDir, ".git")
	_ = os.RemoveAll(gitDir)

	steps := [][]string{
		{"init"},
		{"config", "user.name", "CAKD Platform"},
		{"config", "user.email", "cakd@platform.local"},
		{"add", "."},
		{"commit", "-m", "chore: initial bootstrap by CAKD Platform CLI", "--allow-empty"},
		{"branch", "-M", "main"},
	}

	for _, args := range steps {
		//nolint:gosec
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = cleanDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", errcode.New(errcode.CodeExternal, fmt.Sprintf("local git %s failed: %s", args[0], strings.TrimSpace(string(out))))
		}
	}

	absPath, err := filepath.Abs(cleanDir)
	if err != nil {
		absPath = cleanDir
	}
	repoURL := "file://" + filepath.ToSlash(absPath)

	progress.Emit(ctx, "vcs-local-git", progress.StatusDone, "initialized local git repository at "+repoURL)
	return repoURL, nil
}

// Rollback removes the .git directory created in the local repository.
func (c *Client) Rollback(ctx context.Context, req ports.RepoProvisionRequest) error {
	cleanDir := filepath.Clean(req.SourceDir)
	if strings.Contains(cleanDir, "..") {
		return errcode.New(errcode.CodeValidation, "invalid source directory path")
	}
	gitDir := filepath.Join(cleanDir, ".git")
	return os.RemoveAll(gitDir)
}
