package main

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli"
	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func main() {
	container := composition.NewContainer()
	rootCmd := cli.NewRootCommand(container, os.Stdin, os.Stdout, os.Stderr)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
