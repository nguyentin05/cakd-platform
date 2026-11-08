// Package observe implements the `cakd-cli observe` subcommand.
package observe

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	ucobserve "github.com/nguyentin05/cakd-platform/internal/usecase/observe"
)

// NewCommand creates the `cakd-cli observe` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		filePath  string
		namespace string
		notify    bool
	)
	cmd := &cobra.Command{
		Use:   "observe",
		Short: "Collect Prometheus metrics and Loki logs and run AI root-cause diagnosis",
		RunE: func(cmd *cobra.Command, _ []string) error {
			f.EnsureCredentials(false, false, true)
			rep, err := f.C.ObserveUC.Execute(f.ProgressContext(cmd.Context()), ucobserve.Request{
				SpecPath:     filePath,
				Namespace:    namespace,
				SendToNotify: notify,
			})
			if err != nil {
				return err
			}
			if f.JSON {
				return f.WriteJSON(rep)
			}
			_, _ = fmt.Fprintln(f.Out, "══════════════════════════════════════════")
			_, _ = fmt.Fprintln(f.Out, "CAKD AI DIAGNOSIS")
			_, _ = fmt.Fprintln(f.Out, "══════════════════════════════════════════")
			_, _ = fmt.Fprintln(f.Out, rep.Analysis)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Target Kubernetes namespace")
	cmd.Flags().BoolVar(&notify, "notify", false, "Send diagnosis report to configured Discord webhook")
	return cmd
}
