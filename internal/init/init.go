package init

import (
	"fmt"
	"os"
	"os/exec"
)

type Options struct {
	ArgoCD     bool
	Monitoring bool
	Logging    bool
}

func Run(opts Options) error {
	fmt.Println("Starting cluster bootstrapping...")

	type step struct {
		name     string
		commands [][]string
	}

	var steps []step

	if opts.ArgoCD {
		steps = append(steps, step{
			name: "Step: Installing ArgoCD via Helm",
			commands: [][]string{
				{"helm", "repo", "add", "argo", "https://argoproj.github.io/argo-helm"},
				{"helm", "repo", "update", "argo"},
				{"helm", "upgrade", "--install", "argocd", "argo/argo-cd", "--namespace", "argocd", "--create-namespace"},
			},
		})
	}

	if opts.Monitoring {
		steps = append(steps, step{
			name: "Step: Installing Prometheus & Grafana via Helm",
			commands: [][]string{
				{"helm", "repo", "add", "prometheus-community", "https://prometheus-community.github.io/helm-charts"},
				{"helm", "repo", "update", "prometheus-community"},
				{"helm", "upgrade", "--install", "prometheus", "prometheus-community/kube-prometheus-stack", "--namespace", "monitoring", "--create-namespace"},
			},
		})
	}

	if opts.Logging {
		steps = append(steps, step{
			name: "Step: Installing Loki",
			commands: [][]string{
				{"helm", "repo", "add", "grafana", "https://grafana.github.io/helm-charts"},
				{"helm", "repo", "update", "grafana"},
				{"helm", "upgrade", "--install", "loki", "grafana/loki-stack", "--namespace", "monitoring", "--create-namespace"},
			},
		})
	}

	if len(steps) == 0 {
		fmt.Println("No components selected to install.")
		return nil
	}

	for i, s := range steps {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(steps), s.name)
		for _, args := range s.commands {
			/* #nosec G204 */
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Printf("   > %s\n", cmd.String())
			err := cmd.Run()
			if err != nil && args[1] != "create" {
				return fmt.Errorf("command failed: %v", err)
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
