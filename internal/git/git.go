package git

import (
	"fmt"
	"os"
	"os/exec"
)

// InitAndPush initializes a git repository in projectDir and pushes to the remote
func InitAndPush(projectDir, repoCloneURL, ghToken string) error {
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

	// Construct authenticated remote URL: https://<token>@github.com/owner/repo.git
	authedURL := insertTokenInURL(repoCloneURL, ghToken)

	if err := run(projectDir, "remote", "add", "origin", authedURL); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	if err := run(projectDir, "push", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	// Security: replace authenticated URL with clean URL in .git/config
	// to avoid persisting the token in plaintext
	if err := run(projectDir, "remote", "set-url", "origin", repoCloneURL); err != nil {
		return fmt.Errorf("git remote set-url failed: %w", err)
	}

	return nil
}

// insertTokenInURL converts https://github.com/owner/repo.git
// into https://<token>@github.com/owner/repo.git
func insertTokenInURL(cloneURL, token string) string {
	// cloneURL format: https://github.com/owner/repo.git
	const prefix = "https://"
	if len(cloneURL) > len(prefix) {
		return prefix + token + "@" + cloneURL[len(prefix):]
	}
	return cloneURL
}

func run(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
