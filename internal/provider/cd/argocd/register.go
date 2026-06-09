package argocd

import (
	"fmt"
	"os/exec"
)

func Register(manifestPath string) error {
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
