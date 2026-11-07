package credstore

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/platform/redact"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

//nolint:gosec
const (
	envGitHubToken      = "GITHUB_TOKEN"
	envCAKDGitHubToken  = "CAKD_GITHUB_TOKEN"
	envDiscordToken     = "DISCORD_BOT_TOKEN"
	envCAKDDiscordToken = "CAKD_DISCORD_BOT_TOKEN"
	envDiscordGuild     = "DISCORD_GUILD_ID"
	envCAKDDiscordGuild = "CAKD_DISCORD_GUILD_ID"
	envGeminiKey        = "GEMINI_API_KEY"
	envCAKDGeminiKey    = "CAKD_GEMINI_API_KEY"
)

type fileSchema struct {
	Github struct {
		Token string `yaml:"token,omitempty"`
	} `yaml:"github,omitempty"`
	Discord struct {
		Token   string `yaml:"token,omitempty"`
		GuildID string `yaml:"guild_id,omitempty"`
	} `yaml:"discord,omitempty"`
	Gemini struct {
		APIKey string `yaml:"api_key,omitempty"`
	} `yaml:"gemini,omitempty"`
}

// Store implements ports.CredentialStore backed by environment variables and ~/.cakd/credentials.yaml.
type Store struct {
	path string
}

// NewStore creates a credential Store using the default ~/.cakd/credentials.yaml path.
func NewStore() *Store {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Store{path: ".cakd/credentials.yaml"}
	}
	return &Store{path: filepath.Join(home, ".cakd", "credentials.yaml")}
}

// NewStoreWithPath creates a credential Store using a custom file path.
func NewStoreWithPath(path string) *Store {
	return &Store{path: path}
}

// Path returns the configured credentials file path.
func (s *Store) Path() string {
	return s.path
}

func (s *Store) loadFile() (fileSchema, error) {
	var raw fileSchema
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return raw, nil
		}
		return raw, errcode.Wrap(errcode.CodeExternal, "failed to read credentials file", err)
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return raw, errcode.Wrap(errcode.CodeValidation, "invalid credentials YAML", err)
	}
	return raw, nil
}

func (s *Store) writeFile(raw fileSchema) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to create credentials directory", err)
	}
	data, err := yaml.Marshal(&raw)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to serialize credentials", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to write credentials file", err)
	}
	return nil
}

// Get resolves credentials from environment variables first, then falls back to the local credentials file.
func (s *Store) Get() (ports.CredentialSet, error) {
	raw, err := s.loadFile()
	if err != nil {
		return ports.CredentialSet{}, err
	}

	out := ports.CredentialSet{
		GitHubToken:     firstNonEmpty(os.Getenv(envCAKDGitHubToken), os.Getenv(envGitHubToken), raw.Github.Token),
		DiscordBotToken: firstNonEmpty(os.Getenv(envCAKDDiscordToken), os.Getenv(envDiscordToken), raw.Discord.Token),
		DiscordGuildID:  firstNonEmpty(os.Getenv(envCAKDDiscordGuild), os.Getenv(envDiscordGuild), raw.Discord.GuildID),
		GeminiAPIKey:    firstNonEmpty(os.Getenv(envCAKDGeminiKey), os.Getenv(envGeminiKey), raw.Gemini.APIKey),
	}
	return out, nil
}

// Save merges non-empty fields from creds into the local credentials file.
func (s *Store) Save(creds ports.CredentialSet) error {
	raw, err := s.loadFile()
	if err != nil {
		return err
	}
	if creds.GitHubToken != "" {
		raw.Github.Token = creds.GitHubToken
	}
	if creds.DiscordBotToken != "" {
		raw.Discord.Token = creds.DiscordBotToken
	}
	if creds.DiscordGuildID != "" {
		raw.Discord.GuildID = creds.DiscordGuildID
	}
	if creds.GeminiAPIKey != "" {
		raw.Gemini.APIKey = creds.GeminiAPIKey
	}
	return s.writeFile(raw)
}

// Delete removes stored credentials for a specific provider ("github", "discord", "gemini", or "all").
func (s *Store) Delete(provider string) error {
	raw, err := s.loadFile()
	if err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "github":
		raw.Github.Token = ""
	case "discord":
		raw.Discord.Token = ""
		raw.Discord.GuildID = ""
	case "gemini":
		raw.Gemini.APIKey = ""
	case "all", "":
		raw = fileSchema{}
	default:
		return errcode.New(errcode.CodeValidation, "unknown credential provider: "+provider)
	}
	return s.writeFile(raw)
}

// Status returns a masked summary of currently configured credentials.
func (s *Store) Status() (ports.CredentialStatus, error) {
	creds, err := s.Get()
	if err != nil {
		return ports.CredentialStatus{}, err
	}
	return ports.CredentialStatus{
		GitHubToken:     redact.MaskSecret(creds.GitHubToken),
		DiscordBotToken: redact.MaskSecret(creds.DiscordBotToken),
		DiscordGuildID:  creds.DiscordGuildID,
		GeminiAPIKey:    redact.MaskSecret(creds.GeminiAPIKey),
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
