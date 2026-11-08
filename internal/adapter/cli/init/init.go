// Package initcmd implements the `cakd-cli init` subcommand.
// The directory is named `init`, but Go forbids `init` as a package name.
package initcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/usecase/initcluster"
)

// NewCommand creates the `cakd-cli init` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		filePath   string
		argocd     bool
		monitoring bool
		logging    bool
		installAll bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Kubernetes cluster (local or AWS EC2) and install platform add-ons",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if installAll {
				argocd = true
				monitoring = true
				logging = true
			}
			specPath := ""
			if cmd.Flags().Changed("file") {
				specPath = filePath
			} else if _, err := os.Stat(filePath); err == nil {
				specPath = filePath
			}

			if monitoring {
				f.EnsureCredentials(false, true, true)
			}

			st, err := f.C.InitUC.Execute(f.ProgressContext(cmd.Context()), initcluster.Request{
				SpecPath:   specPath,
				ArgoCD:     argocd,
				Monitoring: monitoring,
				Logging:    logging,
			})
			if err != nil {
				return err
			}
			if f.JSON {
				return f.WriteJSON(st)
			}
			_, _ = fmt.Fprintf(f.Out, "✔ Cluster initialized (provider: %s, domain: %s)\n", st.Provider, st.Domain)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	cmd.Flags().BoolVar(&argocd, "argocd", false, "Install ArgoCD")
	cmd.Flags().BoolVar(&monitoring, "monitoring", false, "Install Prometheus Stack and CAKD Agent")
	cmd.Flags().BoolVar(&logging, "logging", false, "Install Grafana Loki and Promtail")
	cmd.Flags().BoolVar(&installAll, "all", false, "Install all cluster add-ons (ArgoCD, Monitoring, Logging)")
	return cmd
}
