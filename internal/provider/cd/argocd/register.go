package argocd

import (
	"fmt"
	"os/exec"
)

// ArgoCD implements the [cd.CD] interface for ArgoCD GitOps registration.
type ArgoCD struct{}

// New initializes and returns a new ArgoCD client instance.
func New() *ArgoCD {
	return &ArgoCD{}
}

// Register registers a new GitOps application with the local Kubernetes ArgoCD instance
// by applying the generated application manifest using the kubectl CLI.
func (a *ArgoCD) Register(manifestPath string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl is not installed or not in PATH. Please install it to register ArgoCD applications")
	}

	cmd := exec.Command("kubectl", "apply", "-f", manifestPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %s", string(output))
	}

	return nil
}
