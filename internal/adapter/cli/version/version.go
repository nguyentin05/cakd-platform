// Package version implements the `cakd-cli version` subcommand.
package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/platform/buildinfo"
)

// NewCommand creates the `cakd-cli version` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CAKD Platform CLI version",
		RunE: func(_ *cobra.Command, _ []string) error {
			if f.JSON {
				return f.WriteJSON(map[string]string{"version": buildinfo.Version})
			}
			_, _ = fmt.Fprintf(f.Out, "cakd-cli version %s\n", buildinfo.Version)
			return nil
		},
	}
}
