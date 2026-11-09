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
				return runCreateWizard(cmd, f, filePath, force)
			}

			vcsToken := resolveGitHubAuth(f, filePath, scaffoldOnly)
			res, err := f.C.CreateUC.Execute(f.ProgressContext(cmd.Context()), createproject.Request{
				SpecPath:     filePath,
				Force:        force,
				ScaffoldOnly: scaffoldOnly,
				VCSToken:     vcsToken,
			})
			if err != nil {
				return err
			}
			return printCreateResult(f, res)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite output directory if it already exists")
	cmd.Flags().BoolVar(&scaffoldOnly, "scaffold-only", false, "Generate project files locally without cloud/VCS provisioning")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Launch interactive TUI wizard to build platform.yaml")
	return cmd
}

func runCreateWizard(cmd *cobra.Command, f *cmdutil.Factory, filePath string, force bool) error {
	return tui.RunWizard(cmd.Context(), f.In, f.Out, &tui.Services{
		ValidateUC: f.C.ValidateUC,
		SpecRepo:   f.C.SpecRepo,
		CreateUC:   f.C.CreateUC,
		CatalogUC:  f.C.CatalogUC,
		CredStore:  f.C.CredStore,
	}, tui.Options{
		OutputPath: filePath,
		Force:      force,
	})
}

func resolveGitHubAuth(f *cmdutil.Factory, filePath string, scaffoldOnly bool) string {
	if scaffoldOnly {
		return ""
	}

	needGitHub := true
	if f.C.SpecRepo != nil {
		if proj, err := f.C.SpecRepo.LoadFile(filePath); err == nil && proj != nil && proj.Spec.Providers != nil {
			vcs := proj.Spec.Providers.VersionControl
			if vcs == "local" || vcs == "local-git" {
				needGitHub = false
			}
		}
	}

	if needGitHub {
		f.EnsureCredentials(true, false, false)
		if f.C.CredStore != nil {
			if creds, err := f.C.CredStore.Get(); err == nil {
				return creds.GitHubToken
			}
		}
	}
	return ""
}

func printCreateResult(f *cmdutil.Factory, res createproject.Result) error {
	if f.JSON {
		return f.WriteJSON(res)
	}
	_, _ = fmt.Fprintf(f.Out, "✔ Project %q created at %s (domain: %s)\n", res.ProjectName, res.OutputDir, res.Domain)
	if res.RepoURL != "" {
		_, _ = fmt.Fprintf(f.Out, "  Repository: %s\n", res.RepoURL)
	}
	return nil
}
