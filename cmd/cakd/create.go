package main

import (
	"fmt"

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
			return fmt.Errorf("configuration error: %w", err)
		}

		if err := pipeline.Execute(cfg, forceCreate); err != nil {
			return fmt.Errorf("create failed: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createFilePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	createCmd.Flags().BoolVarP(&forceCreate, "force", "", false, "Force overwrite existing project directory")
}
