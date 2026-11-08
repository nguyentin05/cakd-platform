// Package cmdutil provides shared dependency wiring and helpers for cakd-cli subcommands,
// following the `cmdutil.Factory` pattern used by kubectl and the GitHub CLI.
package cmdutil

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/agenttools"
	"github.com/nguyentin05/cakd-platform/internal/usecase/auth"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
	"github.com/nguyentin05/cakd-platform/internal/usecase/initcluster"
	"github.com/nguyentin05/cakd-platform/internal/usecase/observe"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

// Services bundles application use-case services needed by CLI commands.
type Services struct {
	ValidateUC *validate.Service
	InitUC     *initcluster.Service
	CreateUC   *createproject.Service
	ObserveUC  *observe.Service
	AuthUC     *auth.Service
	AgentTools *agenttools.Catalog
	SpecRepo   ports.SpecRepository
	CredStore  ports.CredentialStore
	StateStore ports.ClusterStateStore
}

// Factory holds decoupled use-case services, IO streams, and global flags shared by all subcommands.
type Factory struct {
	C      *Services
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
	JSON   bool
}

// NewFactory wires a Factory against the real process streams with empty services.
func NewFactory() *Factory {
	return NewFactoryWith(nil, os.Stdin, os.Stdout, os.Stderr)
}

// NewFactoryWith builds a Factory with injected dependencies; nil streams fall back to os streams.
func NewFactoryWith(s *Services, in io.Reader, out, errOut io.Writer) *Factory {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if s == nil {
		s = &Services{}
	}
	return &Factory{C: s, In: in, Out: out, ErrOut: errOut}
}

// BindFlags registers global persistent flags (currently `--json`).
func (f *Factory) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&f.JSON, "json", false, "Format command output as structured JSON")
}

// WriteJSON writes v as indented JSON to the output stream.
func (f *Factory) WriteJSON(v any) error {
	enc := json.NewEncoder(f.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ProgressContext attaches a human-readable progress sink unless JSON output is requested.
func (f *Factory) ProgressContext(ctx context.Context) context.Context {
	if f.JSON {
		return ctx
	}
	return progress.WithSink(ctx, progress.SinkFunc(func(ev progress.Event) {
		_, _ = fmt.Fprintf(f.Out, "  • [%s] %s\n", ev.Phase, ev.Message)
	}))
}

// EnsureCredentials prompts for missing provider credentials when stdin is a terminal.
//
//nolint:gocyclo
func (f *Factory) EnsureCredentials(needGitHub, needDiscord, needGemini bool) {
	file, ok := f.In.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return
	}
	creds, err := f.C.AuthUC.Get()
	if err != nil {
		return
	}
	out := f.Out
	scanner := bufio.NewScanner(f.In)
	updated := false

	if needGitHub && creds.GitHubToken == "" {
		_, _ = fmt.Fprint(out, "Enter GitHub Token: ")
		if b, err := term.ReadPassword(int(file.Fd())); err == nil {
			_, _ = fmt.Fprintln(out)
			creds.GitHubToken = strings.TrimSpace(string(b))
			updated = creds.GitHubToken != ""
		}
	}
	if needDiscord && (creds.DiscordBotToken == "" || creds.DiscordGuildID == "") {
		if creds.DiscordBotToken == "" {
			_, _ = fmt.Fprint(out, "Enter Discord Bot Token: ")
			if b, err := term.ReadPassword(int(file.Fd())); err == nil {
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
		if b, err := term.ReadPassword(int(file.Fd())); err == nil {
			_, _ = fmt.Fprintln(out)
			creds.GeminiAPIKey = strings.TrimSpace(string(b))
			updated = creds.GeminiAPIKey != ""
		}
	}

	if updated {
		_ = f.C.AuthUC.Set(creds)
	}
}
