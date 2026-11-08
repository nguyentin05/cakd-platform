// Package cli is the command-line adapter for CAKD. It builds the Cobra command tree
// and delegates each subcommand to its own package under internal/adapter/cli/.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/auth"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/create"
	initcmd "github.com/nguyentin05/cakd-platform/internal/adapter/cli/init"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/mcp"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/observe"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/schema"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/ui"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/validate"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/version"
)

// NewCakdCmd builds the root `cakd-cli` command and attaches all subcommands.
// A fresh command tree is returned on every call so tests can execute it in isolation.
func NewCakdCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cakd-cli",
		Short:         "CAKD — Cloud Application Kubernetes Deployment Platform",
		Long:          "CAKD Platform CLI, TUI, and MCP Server for declarative cloud-native microservice delivery.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetIn(f.In)
	cmd.SetOut(f.Out)
	cmd.SetErr(f.ErrOut)
	f.BindFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		validate.NewCommand(f),
		initcmd.NewCommand(f),
		create.NewCommand(f),
		observe.NewCommand(f),
		auth.NewCommand(f),
		schema.NewCommand(f),
		ui.NewCommand(f),
		mcp.NewCommand(f),
		version.NewCommand(f),
	)
	return cmd
}

// Execute runs the root command with a signal-aware context: SIGINT/SIGTERM cancels
// in-flight work and exits 130; any other failure prints "Error: <message>" and exits 1.
func Execute(services ...*cmdutil.Services) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var s *cmdutil.Services
	if len(services) > 0 && services[0] != nil {
		s = services[0]
	}
	f := cmdutil.NewFactoryWith(s, os.Stdin, os.Stdout, os.Stderr)
	err := NewCakdCmd(f).ExecuteContext(ctx)
	switch {
	case err == nil:
		return
	case errors.Is(err, context.Canceled):
		_, _ = fmt.Fprintln(os.Stderr, "Canceled.")
		os.Exit(130)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
