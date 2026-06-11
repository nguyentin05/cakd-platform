package main

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/pipeline"
	"github.com/spf13/cobra"
)

var (
	createFilePath string
	forceCreate    bool
)

// createCmd defines the 'create' CLI command which parses a platform config file and
// executes the project bootstrapping/provisioning steps.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloud-native project from platform.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Parse(createFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(1)
		}

		if err := pipeline.Execute(cfg, forceCreate); err != nil {
			fmt.Fprintf(os.Stderr, "Create failed: %v\n", err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createFilePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	createCmd.Flags().BoolVarP(&forceCreate, "force", "", false, "Force overwrite existing project directory")
}
