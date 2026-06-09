package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	cmdAdd              = "add"
	cmdUpdate           = "update"
	cmdUpgrade          = "upgrade"
	flagCreateNamespace = "--create-namespace"
	cmdHelm             = "helm"
	cmdRepo             = "repo"
	flagInstall         = "--install"
	flagNamespace       = "--namespace"
	nsMonitoring        = "monitoring"
)

type Options struct {
	ArgoCD     bool
	Monitoring bool
	Logging    bool
}

//nolint:gocyclo,gosec // intentional wrapper around internal commands
func RunInit(opts Options) error {
	fmt.Println("Starting CAKD Platform Initialization...")

	type step struct {
		name     string
		commands [][]string
	}

	var steps []step

	if opts.ArgoCD {
		steps = append(steps, step{
			name: "Step: Installing ArgoCD via Helm",
			commands: [][]string{
				{cmdHelm, cmdRepo, cmdAdd, "argo", "https://argoproj.github.io/argo-helm"},
				{cmdHelm, cmdRepo, cmdUpdate, "argo"},
				{cmdHelm, cmdUpgrade, flagInstall, "argocd", "argo/argo-cd", flagNamespace, "argocd", flagCreateNamespace},
			},
		})
	}

	var promValuesFile string
	if opts.Monitoring {
		promValuesContent := `alertmanager:
  config:
    route:
      group_by: ['namespace', 'alertname']
      group_wait: 5s
      group_interval: 1m
      repeat_interval: 1h
      receiver: 'cakd-agent'
    receivers:
    - name: 'cakd-agent'
      webhook_configs:
      - url: 'http://cakd-agent.monitoring.svc.cluster.local:8080/api/v1/alerts'
        send_resolved: true`
		tmpFile, err := os.CreateTemp("", "prometheus-values-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		if _, err := tmpFile.WriteString(promValuesContent); err != nil {
			return fmt.Errorf("failed to write to temp file: %w", err)
		}
		tmpFile.Close()
		promValuesFile = tmpFile.Name()
		defer os.Remove(promValuesFile)

		steps = append(steps, step{
			name: "Step: Installing Prometheus Stack via Helm",
			commands: [][]string{
				{cmdHelm, cmdRepo, cmdAdd, "prometheus-community", "https://prometheus-community.github.io/helm-charts"},
				{cmdHelm, cmdRepo, cmdUpdate, "prometheus-community"},
				{cmdHelm, cmdUpgrade, flagInstall, nsMonitoring, "prometheus-community/kube-prometheus-stack", flagNamespace, nsMonitoring, flagCreateNamespace, "-f", promValuesFile},
			},
		})

		steps = append(steps, step{
			name: "Step: Deploying CAKD Agent to Cluster",
			commands: [][]string{
				{"internal-agent-deploy"},
			},
		})
	}

	if opts.Logging {
		steps = append(steps, step{
			name: "Step: Installing Loki",
			commands: [][]string{
				{cmdHelm, cmdRepo, cmdAdd, "grafana", "https://grafana.github.io/helm-charts"},
				{cmdHelm, cmdRepo, cmdUpdate, "grafana"},
				{cmdHelm, cmdUpgrade, flagInstall, "loki", "grafana/loki-stack", flagNamespace, nsMonitoring, flagCreateNamespace},
			},
		})
	}

	if len(steps) == 0 {
		fmt.Println("No components selected to install.")
		return nil
	}

	for i, s := range steps {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(steps), s.name)
		for _, cmdArgs := range s.commands {
			if len(cmdArgs) == 1 && cmdArgs[0] == "internal-agent-deploy" {
				if err := deployAgent(); err != nil {
					return err
				}
				continue
			}

			fmt.Printf("   > %v\n", cmdArgs)
			//nolint:gosec // intentional wrapper around internal commands
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed: %v", cmdArgs)
			}
		}
	}

	fmt.Println("\nCluster bootstrapping completed successfully!")
	if opts.ArgoCD {
		fmt.Println("   - ArgoCD is running in 'argocd' namespace.")
	}
	if opts.Monitoring {
		fmt.Println("   - Prometheus Operator is installed.")
	}
	if opts.Logging {
		fmt.Println("   - Loki is running in 'loki' namespace.")
	}
	return nil
}
