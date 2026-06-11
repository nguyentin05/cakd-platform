package main

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/spf13/cobra"
)

var validateFilePath string

// validateCmd defines the 'validate' CLI command which parses and evaluates the structural
// and business logic rules of a platform.yaml file.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a platform.yaml file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Parse(validateFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Valid! Project: %s | Owner: %s | Language: %s\n",
			cfg.Metadata.Name, cfg.Metadata.Owner, cfg.Services[0].Language)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&validateFilePath, "file", "f", "platform.yaml", "Path to platform.yaml")
}
