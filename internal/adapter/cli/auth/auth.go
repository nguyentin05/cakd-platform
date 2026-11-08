// Package auth implements the `cakd-cli auth` subcommand group (set, status, logout).
package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// NewCommand creates the `cakd-cli auth` subcommand group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage local provider credentials (~/.cakd/credentials.yaml)",
	}
	cmd.AddCommand(newSetCmd(f), newStatusCmd(f), newLogoutCmd(f))
	return cmd
}

func newSetCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		ghToken      string
		discordToken string
		discordGuild string
		geminiKey    string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Save provider credentials locally",
		RunE: func(_ *cobra.Command, _ []string) error {
			if ghToken == "" && discordToken == "" && discordGuild == "" && geminiKey == "" {
				f.EnsureCredentials(true, true, true)
				return nil
			}
			if err := f.C.AuthUC.Set(ports.CredentialSet{
				GitHubToken:     ghToken,
				DiscordBotToken: discordToken,
				DiscordGuildID:  discordGuild,
				GeminiAPIKey:    geminiKey,
			}); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(f.Out, "✔ Credentials saved to %s\n", f.C.AuthUC.Path())
			return nil
		},
	}
	cmd.Flags().StringVar(&ghToken, "github-token", "", "GitHub Personal Access Token")
	cmd.Flags().StringVar(&discordToken, "discord-token", "", "Discord Bot Token")
	cmd.Flags().StringVar(&discordGuild, "discord-guild", "", "Discord Guild ID")
	cmd.Flags().StringVar(&geminiKey, "gemini-key", "", "Google Gemini API Key")
	return cmd
}

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display masked status of configured credentials",
		RunE: func(_ *cobra.Command, _ []string) error {
			st, err := f.C.AuthUC.Status()
			if err != nil {
				return err
			}
			if f.JSON {
				return f.WriteJSON(st)
			}
			_, _ = fmt.Fprintf(f.Out, "Credentials file: %s\n", f.C.AuthUC.Path())
			_, _ = fmt.Fprintf(f.Out, "  GitHub Token:      %s\n", emptyLabel(st.GitHubToken))
			_, _ = fmt.Fprintf(f.Out, "  Discord Bot Token: %s\n", emptyLabel(st.DiscordBotToken))
			_, _ = fmt.Fprintf(f.Out, "  Discord Guild ID:  %s\n", emptyLabel(st.DiscordGuildID))
			_, _ = fmt.Fprintf(f.Out, "  Gemini API Key:    %s\n", emptyLabel(st.GeminiAPIKey))
			return nil
		},
	}
}

func newLogoutCmd(f *cmdutil.Factory) *cobra.Command {
	var provider string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := f.C.AuthUC.Logout(provider); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(f.Out, "✔ Stored credentials removed")
			return nil
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "all", "Provider to clear: github, discord, gemini, or all")
	return cmd
}

func emptyLabel(s string) string {
	if s == "" {
		return "(not configured)"
	}
	return s
}
