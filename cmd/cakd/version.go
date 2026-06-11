package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

// versionCmd defines the 'version' CLI command which prints the current release version of the CAKD CLI.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CAKD Platform CLI\n")
		fmt.Printf("Version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
