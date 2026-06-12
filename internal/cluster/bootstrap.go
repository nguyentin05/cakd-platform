package cluster

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
)

//go:embed k8s/prometheus-values.yaml
var promValuesContent string

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
	providerArgoCD      = "argocd"
)

// Options specifies which third-party infrastructure and observability components to provision
// during the Kubernetes cluster bootstrap step.
type Options struct {
	ArgoCD       bool
	Monitoring   bool
	Logging      bool
	AgentVersion string
}

// RunInit drives the Kubernetes cluster initialization by configuring Helm repositories
// and deploying ArgoCD, Prometheus Stack, CAKD Agent, and Grafana Loki.
//
//nolint:gocyclo,gosec // intentional wrapper around internal commands
func RunInit(opts Options) error {
	fmt.Println("Starting CAKD Platform Initialization...")

	type step struct {
		name     string
		commands [][]string
		run      func() error
	}

	var steps []step

	if opts.ArgoCD {
		steps = append(steps, step{
			name: "Step: Installing ArgoCD via Helm",
			commands: [][]string{
				{cmdHelm, cmdRepo, cmdAdd, "argo", "https://argoproj.github.io/argo-helm"},
				{cmdHelm, cmdRepo, cmdUpdate, "argo"},
				{cmdHelm, cmdUpgrade, flagInstall, providerArgoCD, "argo/argo-cd", flagNamespace, providerArgoCD, flagCreateNamespace},
			},
		})
	}

	var promValuesFile string
	if opts.Monitoring {
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
			run: func() error {
				return deployAgent(opts.AgentVersion)
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
		if s.run != nil {
			if err := s.run(); err != nil {
				return err
			}
			continue
		}
		for _, cmdArgs := range s.commands {
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
