package terraform

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/platform/redact"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

//go:embed modules/github modules/aws_ec2
var moduleFS embed.FS

// RunnerFunc executes a Terraform CLI command inside dir and returns combined output.
type RunnerFunc func(ctx context.Context, dir string, args ...string) ([]byte, error)

// Provisioner implements ports.CloudHostProvisioner and GitHub repository Terraform provisioning
// with isolated state directories (.platform/terraform/cluster vs .platform/terraform/vcs).
type Provisioner struct {
	run RunnerFunc
}

// NewProvisioner creates a Terraform Provisioner backed by the local terraform binary.
func NewProvisioner() *Provisioner {
	return &Provisioner{run: defaultRunner}
}

// NewProvisionerWithRunner creates a Terraform Provisioner with a custom command runner for testing.
func NewProvisionerWithRunner(fn RunnerFunc) *Provisioner {
	return &Provisioner{run: fn}
}

func defaultRunner(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if len(args) > 0 && args[0] == "output" {
		//nolint:gosec
		cmd := exec.CommandContext(ctx, "terraform", args...)
		cmd.Dir = dir
		return cmd.Output()
	}

	logPath := filepath.Join(dir, "terraform.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open terraform log: %w", err)
	}
	defer logFile.Close()

	_, _ = fmt.Fprintf(logFile, "\n=== terraform %v ===\n", args)

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = dir
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w (see log: %s)", err, logPath)
	}
	return nil, nil
}

// ProvisionHost provisions a single-node AWS EC2 host using the embedded aws_ec2 module in an isolated state directory.
func (p *Provisioner) ProvisionHost(ctx context.Context, req ports.CloudHostProvisionRequest) (ports.CloudHostProvisionResult, error) {
	tfDir := req.WorkDir
	if tfDir == "" {
		tfDir = filepath.Join("out", req.ProjectName, ".platform", "terraform", "cluster")
	}

	progress.Emit(ctx, "cluster-iac", progress.StatusRunning, "copying aws_ec2 terraform module to "+tfDir)
	if err := copyEmbeddedModule("modules/aws_ec2", tfDir); err != nil {
		return ports.CloudHostProvisionResult{}, errcode.Wrap(errcode.CodeProvision, "copy aws_ec2 module", err)
	}

	vars := map[string]any{
		"project_name":  req.ProjectName,
		"region":        req.Region,
		"instance_type": req.InstanceType,
		"disk_size":     req.DiskSize,
	}
	if err := writeTFVars(tfDir, vars); err != nil {
		return ports.CloudHostProvisionResult{}, errcode.Wrap(errcode.CodeProvision, "write cluster tfvars", err)
	}

	progress.Emit(ctx, "cluster-iac", progress.StatusRunning, "running terraform init & apply for cloud host")
	if _, err := p.run(ctx, tfDir, "init"); err != nil {
		return ports.CloudHostProvisionResult{}, errcode.Wrap(errcode.CodeProvision, "terraform init failed", err)
	}
	if _, err := p.run(ctx, tfDir, "apply", "-auto-approve"); err != nil {
		return ports.CloudHostProvisionResult{}, errcode.Wrap(errcode.CodeProvision, "terraform apply failed", err)
	}

	outputs, err := p.readOutputs(ctx, tfDir)
	if err != nil {
		return ports.CloudHostProvisionResult{}, errcode.Wrap(errcode.CodeProvision, "parse terraform outputs", err)
	}

	publicIP := outputs["public_ip"]
	privateKeyPEM := outputs["private_key_pem"]
	if publicIP == "" || privateKeyPEM == "" {
		return ports.CloudHostProvisionResult{}, errcode.New(errcode.CodeProvision, "missing public_ip or private_key_pem in terraform outputs")
	}

	sshKeyPath := filepath.Join(tfDir, "id_rsa")
	if err := os.WriteFile(sshKeyPath, []byte(privateKeyPEM), 0o600); err != nil {
		return ports.CloudHostProvisionResult{}, errcode.Wrap(errcode.CodeProvision, "write ssh private key", err)
	}

	progress.Emit(ctx, "cluster-iac", progress.StatusDone, "cloud host provisioned at "+publicIP)
	return ports.CloudHostProvisionResult{
		PublicIP:   publicIP,
		SSHKeyPath: sshKeyPath,
	}, nil
}

// DestroyHost runs terraform destroy in the specified cluster Terraform state directory.
func (p *Provisioner) DestroyHost(ctx context.Context, workDir string) error {
	_, err := p.run(ctx, workDir, "destroy", "-auto-approve")
	if err != nil {
		return errcode.Wrap(errcode.CodeProvision, "terraform destroy host failed", err)
	}
	return nil
}

// ProvisionGitHubRepo provisions a remote GitHub repository using the embedded github module in an isolated state directory.
func (p *Provisioner) ProvisionGitHubRepo(ctx context.Context, req ports.RepoProvisionRequest) (string, error) {
	if req.Token == "" {
		return "", errcode.New(errcode.CodeAuth, "GitHub token is required to provision repository")
	}
	tfDir := req.StateDir
	if tfDir == "" {
		tfDir = filepath.Join(req.SourceDir, ".platform", "terraform", "vcs")
	}

	if err := copyEmbeddedModule("modules/github", tfDir); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "copy github module", err)
	}

	vars := map[string]any{
		"project_name": req.ProjectName,
		"github_owner": req.Owner,
		"github_token": req.Token,
	}
	if err := writeTFVars(tfDir, vars); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "write vcs tfvars", err)
	}

	if _, err := p.run(ctx, tfDir, "init"); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, redact.Sanitize(err.Error(), req.Token), nil)
	}
	if _, err := p.run(ctx, tfDir, "apply", "-auto-approve"); err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, redact.Sanitize(err.Error(), req.Token), nil)
	}

	outputs, err := p.readOutputs(ctx, tfDir)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeProvision, "parse vcs terraform outputs", err)
	}
	repoURL := outputs["repo_clone_url"]
	if repoURL == "" {
		return "", errcode.New(errcode.CodeProvision, "repo_clone_url output not found")
	}
	return repoURL, nil
}

// DestroyGitHubRepo destroys a previously provisioned GitHub repository when rolling back a failed bootstrap.
func (p *Provisioner) DestroyGitHubRepo(ctx context.Context, req ports.RepoProvisionRequest) error {
	tfDir := req.StateDir
	if tfDir == "" {
		tfDir = filepath.Join(req.SourceDir, ".platform", "terraform", "vcs")
	}
	if _, err := p.run(ctx, tfDir, "destroy", "-auto-approve"); err != nil {
		return errcode.Wrap(errcode.CodeProvision, "terraform destroy vcs failed", err)
	}
	return nil
}

func (p *Provisioner) readOutputs(ctx context.Context, tfDir string) (map[string]string, error) {
	out, err := p.run(ctx, tfDir, "output", "-json")
	if err != nil {
		return nil, err
	}
	var raw map[string]struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	res := make(map[string]string, len(raw))
	for k, v := range raw {
		res[k] = v.Value
	}
	return res, nil
}

func copyEmbeddedModule(modulePath, tfDir string) error {
	if err := os.MkdirAll(tfDir, 0o750); err != nil {
		return err
	}
	entries, err := moduleFS.ReadDir(modulePath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := moduleFS.ReadFile(filepath.Join(modulePath, entry.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(tfDir, entry.Name()), data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func writeTFVars(tfDir string, vars map[string]any) error {
	data, err := json.MarshalIndent(vars, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tfDir, "terraform.tfvars.json"), data, 0o600)
}
