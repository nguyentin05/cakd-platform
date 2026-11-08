// Package create implements the `cakd-cli create` subcommand.
package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/adapter/tui"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
)

// NewCommand creates the `cakd-cli create` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		filePath     string
		force        bool
		scaffoldOnly bool
		interactive  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Scaffold microservices, provision repository, and register GitOps deployment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if interactive {
				return tui.RunWizard(cmd.Context(), f.In, f.Out, f.C, tui.Options{
					OutputPath: filePath,
					Force:      force,
				})
			}

			if !scaffoldOnly {
				f.EnsureCredentials(true, false, false)
			}

			res, err := f.C.CreateUC.Execute(f.ProgressContext(cmd.Context()), createproject.Request{
				SpecPath:     filePath,
				Force:        force,
				ScaffoldOnly: scaffoldOnly,
			})
			if err != nil {
				return err
			}
			if f.JSON {
				return f.WriteJSON(res)
			}
			_, _ = fmt.Fprintf(f.Out, "✔ Project %q created at %s (domain: %s)\n", res.ProjectName, res.OutputDir, res.Domain)
			if res.RepoURL != "" {
				_, _ = fmt.Fprintf(f.Out, "  Repository: %s\n", res.RepoURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite output directory if it already exists")
	cmd.Flags().BoolVar(&scaffoldOnly, "scaffold-only", false, "Generate project files locally without cloud/VCS provisioning")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Launch interactive TUI wizard to build platform.yaml")
	return cmd
}
