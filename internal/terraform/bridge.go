package terraform

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

//go:embed modules/github
var moduleFS embed.FS

// Bridge orchestrates Terraform execution for infrastructure provisioning
type Bridge struct {
	cfg      *config.PlatformConfig
	workDir  string
	ghToken  string
}

// TerraformOutputs holds the parsed outputs from terraform
type TerraformOutputs struct {
	RepoFullName string
	RepoHTMLURL  string
	RepoCloneURL string
}

// New creates a new Terraform Bridge
func New(cfg *config.PlatformConfig, workDir string) *Bridge {
	return &Bridge{
		cfg:     cfg,
		workDir: workDir,
	}
}

// Apply copies the embedded Terraform module to workDir, writes tfvars, and runs terraform apply
func (b *Bridge) Apply() (*TerraformOutputs, error) {
	// 1. Resolve GitHub token from environment
	b.ghToken = os.Getenv("GITHUB_TOKEN")
	if b.ghToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required.\n" +
			"Create a token at: https://github.com/settings/tokens\n" +
			"Required scopes: repo, delete_repo")
	}

	tfDir := filepath.Join(b.workDir, ".platform", "terraform")

	// 2. Copy embedded Terraform module to workDir
	if err := b.copyModule(tfDir); err != nil {
		return nil, fmt.Errorf("copy terraform module: %w", err)
	}

	// 3. Write terraform.tfvars.json
	if err := b.writeTfVars(tfDir); err != nil {
		return nil, fmt.Errorf("write tfvars: %w", err)
	}

	// 4. Run terraform init
	if err := b.runTerraform(tfDir, "init"); err != nil {
		return nil, fmt.Errorf("terraform init: %w", err)
	}

	// 5. Run terraform apply
	if err := b.runTerraform(tfDir, "apply", "-auto-approve"); err != nil {
		return nil, fmt.Errorf("terraform apply: %w", err)
	}

	// 6. Parse outputs
	outputs, err := b.parseOutputs(tfDir)
	if err != nil {
		return nil, fmt.Errorf("parse terraform outputs: %w", err)
	}

	return outputs, nil
}

// Destroy runs terraform destroy to clean up resources
func (b *Bridge) Destroy() error {
	tfDir := filepath.Join(b.workDir, ".platform", "terraform")
	return b.runTerraform(tfDir, "destroy", "-auto-approve")
}

// copyModule extracts the embedded Terraform module files to the target directory
func (b *Bridge) copyModule(tfDir string) error {
	if err := os.MkdirAll(tfDir, 0755); err != nil {
		return err
	}

	entries, err := moduleFS.ReadDir("modules/github")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := moduleFS.ReadFile(filepath.Join("modules/github", entry.Name()))
		if err != nil {
			return fmt.Errorf("read embedded file %s: %w", entry.Name(), err)
		}

		outPath := filepath.Join(tfDir, entry.Name())
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("write file %s: %w", outPath, err)
		}
	}

	return nil
}

// writeTfVars converts PlatformConfig into terraform.tfvars.json
func (b *Bridge) writeTfVars(tfDir string) error {
	vars := map[string]string{
		"project_name": b.cfg.Metadata.Name,
		"github_owner": b.cfg.Metadata.Owner,
		"github_token": b.ghToken,
	}

	data, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(tfDir, "terraform.tfvars.json"), data, 0600)
}

// runTerraform executes a terraform command in the given directory.
// Output is written to a log file to avoid leaking sensitive values (tokens, secrets)
// that Terraform may include in error messages or variable summaries.
func (b *Bridge) runTerraform(tfDir string, args ...string) error {
	logPath := filepath.Join(tfDir, "terraform.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open terraform log: %w", err)
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "\n=== terraform %s ===\n", args)

	cmd := exec.Command("terraform", args...)
	cmd.Dir = tfDir
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v (see log: %s)", err, logPath)
	}
	return nil
}

// parseOutputs reads terraform output and returns structured data
func (b *Bridge) parseOutputs(tfDir string) (*TerraformOutputs, error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = tfDir

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Terraform output -json returns: { "key": { "value": "...", "type": "string" } }
	var raw map[string]struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal terraform output: %w", err)
	}

	return &TerraformOutputs{
		RepoFullName: raw["repo_full_name"].Value,
		RepoHTMLURL:  raw["repo_html_url"].Value,
		RepoCloneURL: raw["repo_clone_url"].Value,
	}, nil
}
