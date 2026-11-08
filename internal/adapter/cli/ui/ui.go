// Package ui implements the `cakd-cli ui` subcommand.
package ui

import (
	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/adapter/tui"
)

// NewCommand creates the `cakd-cli ui` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		outPath string
		force   bool
	)
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Launch interactive TUI wizard to build platform.yaml and bootstrap projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return tui.RunWizard(cmd.Context(), f.In, f.Out, &tui.Services{
				ValidateUC: f.C.ValidateUC,
				SpecRepo:   f.C.SpecRepo,
				CreateUC:   f.C.CreateUC,
			}, tui.Options{
				OutputPath: outPath,
				Force:      force,
			})
		},
	}
	cmd.Flags().StringVarP(&outPath, "output", "o", "platform.yaml", "Output path for generated platform.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite output directory if running create")
	return cmd
}
