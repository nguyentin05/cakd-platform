// Package validate implements the `cakd-cli validate` subcommand.
package validate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
)

// NewCommand creates the `cakd-cli validate` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate platform.yaml syntax, schema enums, defaults, and dependency DAG",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := f.C.SpecRepo.LoadFile(filePath)
			if err != nil {
				return err
			}
			if err := f.C.ValidateUC.ValidateAndDefault(proj); err != nil {
				return err
			}
			if f.JSON {
				return f.WriteJSON(map[string]any{
					"valid":   true,
					"project": proj,
				})
			}
			_, _ = fmt.Fprintf(f.Out, "✔ %s is valid (project: %s, services: %d)\n", filePath, proj.Metadata.Name, len(proj.Spec.Services))
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	return cmd
}
