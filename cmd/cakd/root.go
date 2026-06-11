package main

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the CAKD Platform CLI.
var rootCmd = &cobra.Command{
	Use:   "cakd",
	Short: "CAKD Platform CLI — Cloud-Agnostic Kubernetes Developer Platform",
}

// Execute runs the Cobra CLI command framework, parsing input flags and executing
// the target subcommands. It exits the process with code 1 if an error is encountered.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
