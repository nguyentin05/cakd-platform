package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/nguyentin05/cakd-platform/internal/adapter/mcpserver"
	"github.com/nguyentin05/cakd-platform/internal/adapter/tui"
	"github.com/nguyentin05/cakd-platform/internal/composition"
	"github.com/nguyentin05/cakd-platform/internal/platform/buildinfo"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
	"github.com/nguyentin05/cakd-platform/internal/usecase/initcluster"
	"github.com/nguyentin05/cakd-platform/internal/usecase/observe"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/schema"
)

// NewRootCommand constructs the root Cobra command and all subcommands for the cakd CLI binary.
func NewRootCommand(c *composition.Container, in io.Reader, out, errOut io.Writer) *cobra.Command {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	var jsonOutput bool

	rootCmd := &cobra.Command{
		Use:           "cakd",
		Short:         "CAKD — Cloud Application Kubernetes Deployment Platform",
		Long:          "CAKD Platform CLI, TUI, and MCP Server for declarative cloud-native microservice delivery.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.SetIn(in)
	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Format command output as structured JSON")

	rootCmd.AddCommand(
		newValidateCmd(c, out, &jsonOutput),
		newInitCmd(c, in, out, &jsonOutput),
		newCreateCmd(c, in, out, &jsonOutput),
		newObserveCmd(c, in, out, &jsonOutput),
		newAuthCmd(c, in, out, &jsonOutput),
		newSchemaCmd(out),
		newUICmd(c, in, out),
		newMCPCmd(c, in, out),
		newVersionCmd(out, &jsonOutput),
	)

	return rootCmd
}

func attachProgressSink(ctx context.Context, out io.Writer, jsonOutput bool) context.Context {
	if jsonOutput {
		return ctx
	}
	return progress.WithSink(ctx, progress.SinkFunc(func(ev progress.Event) {
		_, _ = fmt.Fprintf(out, "  • [%s] %s\n", ev.Phase, ev.Message)
	}))
}

func newValidateCmd(c *composition.Container, out io.Writer, jsonOutput *bool) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate platform.yaml syntax, schema enums, defaults, and dependency DAG",
		RunE: func(_ *cobra.Command, _ []string) error {
			proj, err := c.SpecRepo.LoadFile(filePath)
			if err != nil {
				return err
			}
			if err := c.ValidateUC.ValidateAndDefault(proj); err != nil {
				return err
			}
			if *jsonOutput {
				return writeJSON(out, map[string]any{
					"valid":   true,
					"project": proj,
				})
			}
			_, _ = fmt.Fprintf(out, "✔ %s is valid (project: %s, services: %d)\n", filePath, proj.Metadata.Name, len(proj.Spec.Services))
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	return cmd
}

func newInitCmd(c *composition.Container, in io.Reader, out io.Writer, jsonOutput *bool) *cobra.Command {
	var (
		filePath   string
		argocd     bool
		monitoring bool
		logging    bool
		installAll bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Kubernetes cluster (local or AWS EC2) and install platform add-ons",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if installAll {
				argocd = true
				monitoring = true
				logging = true
			}
			specPath := ""
			if cmd.Flags().Changed("file") {
				specPath = filePath
			} else if _, err := os.Stat(filePath); err == nil {
				specPath = filePath
			}

			if monitoring {
				ensureInteractiveCredentials(c, in, out, false, true, true)
			}

			ctx := attachProgressSink(cmd.Context(), out, *jsonOutput)
			st, err := c.InitUC.Execute(ctx, initcluster.Request{
				SpecPath:   specPath,
				ArgoCD:     argocd,
				Monitoring: monitoring,
				Logging:    logging,
			})
			if err != nil {
				return err
			}
			if *jsonOutput {
				return writeJSON(out, st)
			}
			_, _ = fmt.Fprintf(out, "✔ Cluster initialized (provider: %s, domain: %s)\n", st.Provider, st.Domain)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	cmd.Flags().BoolVar(&argocd, "argocd", false, "Install ArgoCD")
	cmd.Flags().BoolVar(&monitoring, "monitoring", false, "Install Prometheus Stack and CAKD Agent")
	cmd.Flags().BoolVar(&logging, "logging", false, "Install Grafana Loki and Promtail")
	cmd.Flags().BoolVar(&installAll, "all", false, "Install all cluster add-ons (ArgoCD, Monitoring, Logging)")
	return cmd
}

func newCreateCmd(c *composition.Container, in io.Reader, out io.Writer, jsonOutput *bool) *cobra.Command {
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
				return tui.RunWizard(cmd.Context(), in, out, c, tui.Options{
					OutputPath: filePath,
					Force:      force,
				})
			}

			if !scaffoldOnly {
				ensureInteractiveCredentials(c, in, out, true, false, false)
			}

			ctx := attachProgressSink(cmd.Context(), out, *jsonOutput)
			res, err := c.CreateUC.Execute(ctx, createproject.Request{
				SpecPath:     filePath,
				Force:        force,
				ScaffoldOnly: scaffoldOnly,
			})
			if err != nil {
				return err
			}
			if *jsonOutput {
				return writeJSON(out, res)
			}
			_, _ = fmt.Fprintf(out, "✔ Project %q created at %s (domain: %s)\n", res.ProjectName, res.OutputDir, res.Domain)
			if res.RepoURL != "" {
				_, _ = fmt.Fprintf(out, "  Repository: %s\n", res.RepoURL)
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

func newObserveCmd(c *composition.Container, in io.Reader, out io.Writer, jsonOutput *bool) *cobra.Command {
	var (
		filePath  string
		namespace string
		notify    bool
	)
	cmd := &cobra.Command{
		Use:   "observe",
		Short: "Collect Prometheus metrics and Loki logs and run AI root-cause diagnosis",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ensureInteractiveCredentials(c, in, out, false, false, true)
			ctx := attachProgressSink(cmd.Context(), out, *jsonOutput)
			rep, err := c.ObserveUC.Execute(ctx, observe.Request{
				SpecPath:     filePath,
				Namespace:    namespace,
				SendToNotify: notify,
			})
			if err != nil {
				return err
			}
			if *jsonOutput {
				return writeJSON(out, rep)
			}
			_, _ = fmt.Fprintln(out, "══════════════════════════════════════════")
			_, _ = fmt.Fprintln(out, "CAKD AI DIAGNOSIS")
			_, _ = fmt.Fprintln(out, "══════════════════════════════════════════")
			_, _ = fmt.Fprintln(out, rep.Analysis)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "platform.yaml", "Path to platform.yaml")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Target Kubernetes namespace")
	cmd.Flags().BoolVar(&notify, "notify", false, "Send diagnosis report to configured Discord webhook")
	return cmd
}

func newAuthCmd(c *composition.Container, in io.Reader, out io.Writer, jsonOutput *bool) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage local provider credentials (~/.cakd/credentials.yaml)",
	}

	var (
		ghToken      string
		discordToken string
		discordGuild string
		geminiKey    string
	)
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Save provider credentials locally",
		RunE: func(_ *cobra.Command, _ []string) error {
			if ghToken == "" && discordToken == "" && discordGuild == "" && geminiKey == "" {
				ensureInteractiveCredentials(c, in, out, true, true, true)
				return nil
			}
			if err := c.AuthUC.Set(ports.CredentialSet{
				GitHubToken:     ghToken,
				DiscordBotToken: discordToken,
				DiscordGuildID:  discordGuild,
				GeminiAPIKey:    geminiKey,
			}); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(out, "✔ Credentials saved to %s\n", c.AuthUC.Path())
			return nil
		},
	}
	setCmd.Flags().StringVar(&ghToken, "github-token", "", "GitHub Personal Access Token")
	setCmd.Flags().StringVar(&discordToken, "discord-token", "", "Discord Bot Token")
	setCmd.Flags().StringVar(&discordGuild, "discord-guild", "", "Discord Guild ID")
	setCmd.Flags().StringVar(&geminiKey, "gemini-key", "", "Google Gemini API Key")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display masked status of configured credentials",
		RunE: func(_ *cobra.Command, _ []string) error {
			st, err := c.AuthUC.Status()
			if err != nil {
				return err
			}
			if *jsonOutput {
				return writeJSON(out, st)
			}
			_, _ = fmt.Fprintf(out, "Credentials file: %s\n", c.AuthUC.Path())
			_, _ = fmt.Fprintf(out, "  GitHub Token:      %s\n", emptyLabel(st.GitHubToken))
			_, _ = fmt.Fprintf(out, "  Discord Bot Token: %s\n", emptyLabel(st.DiscordBotToken))
			_, _ = fmt.Fprintf(out, "  Discord Guild ID:  %s\n", emptyLabel(st.DiscordGuildID))
			_, _ = fmt.Fprintf(out, "  Gemini API Key:    %s\n", emptyLabel(st.GeminiAPIKey))
			return nil
		},
	}

	var provider string
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := c.AuthUC.Logout(provider); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(out, "✔ Stored credentials removed")
			return nil
		},
	}
	logoutCmd.Flags().StringVar(&provider, "provider", "all", "Provider to clear: github, discord, gemini, or all")

	authCmd.AddCommand(setCmd, statusCmd, logoutCmd)
	return authCmd
}

func newSchemaCmd(out io.Writer) *cobra.Command {
	var outPath string
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Export JSON Schema (Draft 2020-12) for platform.yaml",
		RunE: func(_ *cobra.Command, _ []string) error {
			data, err := schema.GenerateJSONSchema()
			if err != nil {
				return err
			}
			if outPath != "" {
				if err := os.WriteFile(outPath, append(data, '\n'), 0o600); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(out, "✔ JSON Schema written to %s\n", outPath)
				return nil
			}
			_, _ = fmt.Fprintln(out, string(data))
			return nil
		},
	}
	cmd.Flags().StringVarP(&outPath, "output", "o", "", "Output file path (defaults to stdout)")
	return cmd
}

func newUICmd(c *composition.Container, in io.Reader, out io.Writer) *cobra.Command {
	var (
		outPath string
		force   bool
	)
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Launch interactive TUI wizard to build platform.yaml and bootstrap projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return tui.RunWizard(cmd.Context(), in, out, c, tui.Options{
				OutputPath: outPath,
				Force:      force,
			})
		},
	}
	cmd.Flags().StringVarP(&outPath, "output", "o", "platform.yaml", "Output path for generated platform.yaml")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite output directory if running create")
	return cmd
}

func newMCPCmd(c *composition.Container, in io.Reader, out io.Writer) *cobra.Command {
	var (
		transport string
		addr      string
		authToken string
	)
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the Model Context Protocol (MCP) server over stdio or Streamable HTTP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			srv := mcpserver.NewServer(c.AgentTools)
			if transport == "http" {
				if authToken == "" {
					authToken = os.Getenv("CAKD_MCP_AUTH_TOKEN")
				}
				httpSrv := &http.Server{
					Addr:              addr,
					Handler:           srv.NewHTTPHandler(authToken),
					ReadHeaderTimeout: 10 * time.Second,
				}
				return httpSrv.ListenAndServe()
			}
			return srv.ServeStdio(cmd.Context(), in, out)
		},
	}
	cmd.Flags().StringVar(&transport, "transport", "stdio", "MCP transport: stdio or http")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8090", "Listen address when --transport=http")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Bearer token for HTTP MCP transport (or CAKD_MCP_AUTH_TOKEN)")
	return cmd
}

func newVersionCmd(out io.Writer, jsonOutput *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CAKD Platform CLI version",
		RunE: func(_ *cobra.Command, _ []string) error {
			if *jsonOutput {
				return writeJSON(out, map[string]string{"version": buildinfo.Version})
			}
			_, _ = fmt.Fprintf(out, "cakd version %s\n", buildinfo.Version)
			return nil
		},
	}
}

func ensureInteractiveCredentials(c *composition.Container, in io.Reader, out io.Writer, needGitHub, needDiscord, needGemini bool) {
	f, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return
	}
	creds, err := c.AuthUC.Get()
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(in)
	updated := false

	if needGitHub && creds.GitHubToken == "" {
		_, _ = fmt.Fprint(out, "Enter GitHub Token: ")
		if b, err := term.ReadPassword(int(f.Fd())); err == nil {
			_, _ = fmt.Fprintln(out)
			creds.GitHubToken = strings.TrimSpace(string(b))
			updated = creds.GitHubToken != ""
		}
	}
	if needDiscord && (creds.DiscordBotToken == "" || creds.DiscordGuildID == "") {
		if creds.DiscordBotToken == "" {
			_, _ = fmt.Fprint(out, "Enter Discord Bot Token: ")
			if b, err := term.ReadPassword(int(f.Fd())); err == nil {
				_, _ = fmt.Fprintln(out)
				creds.DiscordBotToken = strings.TrimSpace(string(b))
				updated = true
			}
		}
		if creds.DiscordGuildID == "" {
			_, _ = fmt.Fprint(out, "Enter Discord Guild ID: ")
			if scanner.Scan() {
				creds.DiscordGuildID = strings.TrimSpace(scanner.Text())
				updated = true
			}
		}
	}
	if needGemini && creds.GeminiAPIKey == "" {
		_, _ = fmt.Fprint(out, "Enter Gemini API Key: ")
		if b, err := term.ReadPassword(int(f.Fd())); err == nil {
			_, _ = fmt.Fprintln(out)
			creds.GeminiAPIKey = strings.TrimSpace(string(b))
			updated = creds.GeminiAPIKey != ""
		}
	}

	if updated {
		_ = c.AuthUC.Set(creds)
	}
}

func writeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func emptyLabel(s string) string {
	if s == "" {
		return "(not configured)"
	}
	return s
}
