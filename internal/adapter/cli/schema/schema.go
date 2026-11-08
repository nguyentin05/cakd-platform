// Package schema implements the `cakd-cli schema` subcommand.
package schema

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	ucschema "github.com/nguyentin05/cakd-platform/internal/usecase/schema"
)

// NewCommand creates the `cakd-cli schema` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outPath string
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Export JSON Schema (Draft 2020-12) for platform.yaml",
		RunE: func(_ *cobra.Command, _ []string) error {
			data, err := ucschema.GenerateJSONSchema()
			if err != nil {
				return err
			}
			if outPath != "" {
				if err := os.WriteFile(outPath, append(data, '\n'), 0o600); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(f.Out, "✔ JSON Schema written to %s\n", outPath)
				return nil
			}
			_, _ = fmt.Fprintln(f.Out, string(data))
			return nil
		},
	}
	cmd.Flags().StringVarP(&outPath, "output", "o", "", "Output file path (defaults to stdout)")
	return cmd
}
