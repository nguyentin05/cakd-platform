package main

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/observe"
	"github.com/nguyentin05/cakd-platform/internal/provider/llm/gemini"
	"github.com/nguyentin05/cakd-platform/internal/provider/logging/loki"
	"github.com/nguyentin05/cakd-platform/internal/provider/monitoring/prometheus"
	"github.com/spf13/cobra"
)

var observeCmd = &cobra.Command{
	Use:   "observe <project-name>",
	Short: "Use AI to observe, diagnose, and troubleshoot a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: GEMINI_API_KEY environment variable is required")
			fmt.Fprintln(os.Stderr, "Please set it using: export GEMINI_API_KEY=\"your_api_key\"")
			os.Exit(1)
		}

		fmt.Printf("Starting AI Observability for project: %s\n", projectName)

		metrics := prometheus.NewClient()
		logs := loki.NewClient()
		ai := gemini.NewClient(apiKey)

		service := observe.NewObserverService(metrics, logs, ai)
		if err := service.Diagnose(projectName); err != nil {
			fmt.Fprintf(os.Stderr, "Observability failed: %v\n", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(observeCmd)
}
