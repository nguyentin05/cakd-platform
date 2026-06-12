package github

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		if err := runGitWithRedaction(projectDir, ghToken, step.args...); err != nil {
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	authedURL := insertTokenInURL(repoCloneURL, ghToken)

	if err := runGitWithRedaction(projectDir, ghToken, "remote", "add", "origin", authedURL); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	// Always ensure remote URL is cleaned up so we never leak the token in .git/config on failure.
	defer func() {
		_ = runGitWithRedaction(projectDir, ghToken, "remote", "set-url", "origin", repoCloneURL)
	}()

	// Force push (-f) is intentional here because this is the initial bootstrap
	// of a newly generated project repository, ensuring origin main is set cleanly.
	if err := runGitWithRedaction(projectDir, ghToken, "push", "-f", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
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

// runGitWithRedaction executes a git command in the specified directory, captures stdout/stderr,
// redacts occurrences of the authentication token, and writes the redacted outputs to os.Stdout/os.Stderr.
func runGitWithRedaction(dir string, token string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if token != "" {
		stdoutStr = strings.ReplaceAll(stdoutStr, token, "******")
		stderrStr = strings.ReplaceAll(stderrStr, token, "******")
	}

	_, _ = os.Stdout.Write([]byte(stdoutStr))
	_, _ = os.Stderr.Write([]byte(stderrStr))

	return err
}
