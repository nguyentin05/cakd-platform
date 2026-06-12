package terraform

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/iac"
	"github.com/nguyentin05/cakd-platform/internal/schema"
)

//go:embed modules/github
var moduleFS embed.FS

func init() {
	iac.NewEngine = func(cfg *schema.PlatformConfig, workDir string) iac.Engine {
		return New(cfg, workDir)
	}
}

// Bridge implements the [iac.Engine] interface using Terraform to provision infrastructure.
type Bridge struct {
	cfg     *schema.PlatformConfig
	workDir string
	ghToken string
}

// New initializes and returns a new Terraform Bridge instance.
func New(cfg *schema.PlatformConfig, workDir string) *Bridge {
	return &Bridge{
		cfg:     cfg,
		workDir: workDir,
	}
}

// Apply copies the embedded Terraform modules, writes the input variables,
// runs 'terraform init' and 'terraform apply', and returns the parsed outputs.
func (b *Bridge) Apply() (map[string]string, error) {
	b.ghToken = config.GetGithubToken()
	if b.ghToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required.\n" +
			"Create a token at: https://github.com/settings/tokens\n" +
			"Required scopes: repo, delete_repo")
	}

	tfDir := filepath.Join(b.workDir, ".platform", "terraform")

	if err := b.copyModule(tfDir); err != nil {
		return nil, fmt.Errorf("copy terraform module: %w", err)
	}

	if err := b.writeTfVars(tfDir); err != nil {
		return nil, fmt.Errorf("write tfvars: %w", err)
	}

	if err := b.runTerraform(tfDir, "init"); err != nil {
		return nil, fmt.Errorf("terraform init: %w", err)
	}

	if err := b.runTerraform(tfDir, "apply", "-auto-approve"); err != nil {
		return nil, fmt.Errorf("terraform apply: %w", err)
	}

	outputs, err := b.parseOutputs(tfDir)
	if err != nil {
		return nil, fmt.Errorf("parse terraform outputs: %w", err)
	}

	return outputs, nil
}

// Destroy runs 'terraform destroy' to teardown all provisioned infrastructure resources.
func (b *Bridge) Destroy() error {
	tfDir := filepath.Join(b.workDir, ".platform", "terraform")
	return b.runTerraform(tfDir, "destroy", "-auto-approve")
}

// copyModule copies the embedded GitHub Terraform module files to the target workspace directory.
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
		if err := os.WriteFile(outPath, data, 0600); err != nil {
			return fmt.Errorf("write file %s: %w", outPath, err)
		}
	}

	return nil
}

// writeTfVars serializes variables to JSON format and writes them to terraform.tfvars.json.
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

// runTerraform executes a Terraform command in the specified directory, appending logs to terraform.log.
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

// parseOutputs queries Terraform output variables in JSON format and parses them into a map.
func (b *Bridge) parseOutputs(tfDir string) (map[string]string, error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = tfDir

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var raw map[string]struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal terraform output: %w", err)
	}

	outputs := make(map[string]string)
	for k, v := range raw {
		outputs[k] = v.Value
	}

	return outputs, nil
}
