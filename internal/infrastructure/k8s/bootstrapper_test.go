package k8s_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/infrastructure/k8s"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

func TestBootstrapperLocalClusterRunsHelmWithScopedEnv(t *testing.T) {
	var recordedCmds []string
	bs := k8s.NewBootstrapperWithExecutor(func(_ context.Context, kubeconfig string, name string, args ...string) ([]byte, error) {
		recordedCmds = append(recordedCmds, kubeconfig+"|"+name+" "+strings.Join(args, " "))
		return []byte("ok"), nil
	})

	state, err := bs.BootstrapCluster(context.Background(), ports.ClusterBootstrapOptions{
		Provider:       "local",
		Domain:         "127.0.0.1.nip.io",
		InstallArgoCD:  true,
		InstallMonitor: true,
		InstallLogging: true,
		InstallAgent:   true,
	})
	if err != nil {
		t.Fatalf("BootstrapCluster failed: %v", err)
	}
	if state.Domain != "127.0.0.1.nip.io" {
		t.Fatalf("expected domain 127.0.0.1.nip.io, got %s", state.Domain)
	}
	if len(recordedCmds) == 0 {
		t.Fatalf("expected helm/kubectl commands to be executed")
	}
}
