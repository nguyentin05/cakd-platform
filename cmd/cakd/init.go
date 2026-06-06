package main

import (
	"fmt"
	"os"

	pkginit "github.com/nguyentin05/cakd-platform/internal/init"
	"github.com/spf13/cobra"
)

var (
	optArgoCD     bool
	optMonitoring bool
	optLogging    bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap the Kubernetes cluster with Platform Infrastructure (ArgoCD, Prometheus, Loki)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !optArgoCD && !optMonitoring && !optLogging {
			optArgoCD = true
			optMonitoring = true
			optLogging = true
		}

		opts := pkginit.Options{
			ArgoCD:     optArgoCD,
			Monitoring: optMonitoring,
			Logging:    optLogging,
		}

		if err := pkginit.Run(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Cluster initialization failed: %v\n", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&optArgoCD, "argocd", false, "Install ArgoCD")
	initCmd.Flags().BoolVar(&optMonitoring, "monitoring", false, "Install Prometheus & Grafana")
	initCmd.Flags().BoolVar(&optLogging, "logging", false, "Install Loki")
}
