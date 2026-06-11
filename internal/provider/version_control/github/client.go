package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Client implements the [version_control.VersionControl] interface for GitHub.
// It manages local git initialization and pushing committed code to GitHub repositories.
type Client struct{}

// NewClient initializes and returns a new GitHub VCS Client.
func NewClient() *Client {
	return &Client{}
}

// InitAndPush initializes a local git repository, commits the generated code files,
// inserts the authentication token into the repository clone URL, and pushes the code to GitHub.
func (c *Client) InitAndPush(projectDir, repoCloneURL, ghToken string) error {
	os.RemoveAll(filepath.Join(projectDir, ".git"))

	steps := []struct {
		name string
		args []string
	}{
		{"git init", []string{"init"}},
		{"git add", []string{"add", "."}},
		{"git commit", []string{"commit", "-m", "chore: initial bootstrap by CAKD Platform CLI"}},
		{"git branch", []string{"branch", "-M", "main"}},
	}

	for _, step := range steps {
		if err := run(projectDir, step.args...); err != nil {
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	authedURL := insertTokenInURL(repoCloneURL, ghToken)

	if err := run(projectDir, "remote", "add", "origin", authedURL); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	if err := run(projectDir, "push", "-f", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	if err := run(projectDir, "remote", "set-url", "origin", repoCloneURL); err != nil {
		return fmt.Errorf("git remote set-url failed: %w", err)
	}

	return nil
}

// insertTokenInURL embeds the personal access token directly into the HTTPS clone URL
// to authenticate push requests.
func insertTokenInURL(cloneURL, token string) string {
	const prefix = "https://"
	if len(cloneURL) > len(prefix) {
		return prefix + token + "@" + cloneURL[len(prefix):]
	}
	return cloneURL
}

// run is a helper that executes a git command in the specified directory,
// redirecting standard outputs and errors.
func run(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
