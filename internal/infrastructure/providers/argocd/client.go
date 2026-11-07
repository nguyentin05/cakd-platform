package argocd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// Client implements ports.GitOpsProvider by applying generated ArgoCD Application manifests via kubectl.
type Client struct{}

// NewClient creates a new ArgoCD GitOps provider Client.
func NewClient() *Client {
	return &Client{}
}

// RegisterApplication applies the generated ArgoCD application manifest to the cluster.
func (c *Client) RegisterApplication(ctx context.Context, req ports.GitOpsRegisterRequest) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return errcode.Wrap(errcode.CodeProvision, "kubectl is not installed or not in PATH", err)
	}

	manifestPath := req.TargetPath
	if manifestPath == "" {
		manifestPath = filepath.Join("out", req.AppName, "deploy", "application.yaml")
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", manifestPath)
	if req.KubeconfigPath != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+req.KubeconfigPath)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errcode.New(errcode.CodeProvision, fmt.Sprintf("kubectl apply failed: %s", strings.TrimSpace(string(out))))
	}
	return nil
}
