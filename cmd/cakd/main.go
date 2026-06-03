package main

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/create"
	"github.com/spf13/cobra"
)

var version = "v1.0.0-dev"

func main() {
	var filePath string

	rootCmd := &cobra.Command{
		Use:   "cakd",
		Short: "CAKD Platform CLI — Cloud-Agnostic Kubernetes Developer Platform",
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a platform.yaml file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Parse(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Valid! Project: %s | Owner: %s | Language: %s\n",
				cfg.Metadata.Name, cfg.Metadata.Owner, cfg.Spec.Language)
			return nil
		},
	}
	validateCmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a new cloud-native project from platform.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Parse(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
				os.Exit(1)
			}

			if err := create.Run(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Bootstrap failed: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}
	createCmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cakd version %s\n", version)
		},
	}

	rootCmd.AddCommand(validateCmd, createCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
