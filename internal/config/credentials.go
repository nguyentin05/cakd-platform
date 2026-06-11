package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Credentials represents the structured YAML config for local auth tokens.
type Credentials struct {
	Github struct {
		Token string `yaml:"token"`
	} `yaml:"github"`
	Discord struct {
		Token   string `yaml:"token"`
		GuildID string `yaml:"guild_id"`
	} `yaml:"discord"`
	Gemini struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"gemini"`
}

//nolint:gosec
const (
	githubTokenEnv      = "GITHUB_TOKEN"
	cakdGithubTokenEnv  = "CAKD_GITHUB_TOKEN"
	discordTokenEnv     = "DISCORD_BOT_TOKEN"
	cakdDiscordTokenEnv = "CAKD_DISCORD_BOT_TOKEN"
	discordGuildEnv     = "DISCORD_GUILD_ID"
	cakdDiscordGuildEnv = "CAKD_DISCORD_GUILD_ID"
	geminiKeyEnv        = "GEMINI_API_KEY"
	cakdGeminiKeyEnv    = "CAKD_GEMINI_API_KEY"
)

var credentialsPathOverride string

// GetCredentialsPath returns the absolute path to ~/.cakd/credentials.yaml.
func GetCredentialsPath() (string, error) {
	if credentialsPathOverride != "" {
		return credentialsPathOverride, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cakd", "credentials.yaml"), nil
}

// LoadCredentials reads and parses ~/.cakd/credentials.yaml.
func LoadCredentials() (*Credentials, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, err
	}
	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// SaveCredentials writes the credentials back to ~/.cakd/credentials.yaml with 0600 permissions.
func SaveCredentials(creds *Credentials) error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// isInteractive checks if stdin is a terminal.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptSecure prompts the user for a masked input on the terminal.
func promptSecure(prompt string) (string, error) {
	fmt.Print(prompt)
	if !isInteractive() {
		return "", fmt.Errorf("stdin is not a terminal, cannot prompt securely")
	}
	byteVal, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(byteVal)), nil
}

// promptPlain prompts the user for a visible text input.
func promptPlain(prompt string) (string, error) {
	fmt.Print(prompt)
	var val string
	_, err := fmt.Scanln(&val)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

// promptYesNo asks a Yes/No question, defaulting to Yes on empty input.
func promptYesNo(prompt string) bool {
	fmt.Print(prompt)
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes"
}

// GetGithubToken returns the GitHub Token resolving through the 3 layers.
func GetGithubToken() string {
	// Layer 1: Env vars
	if val := os.Getenv(cakdGithubTokenEnv); val != "" {
		return val
	}
	if val := os.Getenv(githubTokenEnv); val != "" {
		return val
	}

	// Layer 2: Credentials file
	creds, err := LoadCredentials()
	if err == nil && creds.Github.Token != "" {
		return creds.Github.Token
	}

	// Layer 3: Interactive CLI Prompt
	if isInteractive() {
		fmt.Println("⚠️  GitHub Token not found in environment or local settings.")
		token, err := promptSecure("Enter GitHub Token: ")
		if err == nil && token != "" {
			if promptYesNo("Save this token to ~/.cakd/credentials.yaml? [Y/n]: ") {
				creds, err = LoadCredentials()
				if err == nil {
					creds.Github.Token = token
					_ = SaveCredentials(creds)
					fmt.Println("✅ Saved GitHub token locally.")
				}
			}
			return token
		}
	}

	return ""
}

// GetDiscordCredentials returns the Discord Token and Guild ID resolving through the 3 layers.
func GetDiscordCredentials() (token string, guildID string, err error) {
	// Layer 1: Env vars
	tok := os.Getenv(cakdDiscordTokenEnv)
	if tok == "" {
		tok = os.Getenv(discordTokenEnv)
	}
	gid := os.Getenv(cakdDiscordGuildEnv)
	if gid == "" {
		gid = os.Getenv(discordGuildEnv)
	}

	if tok != "" && gid != "" {
		return tok, gid, nil
	}

	// Layer 2: Credentials file
	creds, err := LoadCredentials()
	if err == nil {
		if tok == "" {
			tok = creds.Discord.Token
		}
		if gid == "" {
			gid = creds.Discord.GuildID
		}
	}

	if tok != "" && gid != "" {
		return tok, gid, nil
	}

	// Layer 3: Interactive CLI Prompt
	if isInteractive() {
		fmt.Println("⚠️  Discord Bot Token or Guild ID not found in environment or local settings.")
		return promptDiscordCredentials(tok, gid)
	}

	return "", "", fmt.Errorf("discord credentials are not fully configured")
}

// promptDiscordCredentials handles the interactive prompt step for Discord setup.
func promptDiscordCredentials(tok, gid string) (string, string, error) {
	var err error
	if tok == "" {
		tok, err = promptSecure("Enter Discord Bot Token: ")
		if err != nil || tok == "" {
			return "", "", fmt.Errorf("canceled or invalid bot token: %w", err)
		}
	}
	if gid == "" {
		gid, err = promptPlain("Enter Discord Guild ID: ")
		if err != nil || gid == "" {
			return "", "", fmt.Errorf("canceled or invalid guild ID: %w", err)
		}
	}

	if promptYesNo("Save Discord credentials to ~/.cakd/credentials.yaml? [Y/n]: ") {
		creds, err := LoadCredentials()
		if err == nil {
			creds.Discord.Token = tok
			creds.Discord.GuildID = gid
			_ = SaveCredentials(creds)
			fmt.Println("✅ Saved Discord credentials locally.")
		}
	}
	return tok, gid, nil
}

// GetGeminiAPIKey returns the Gemini API Key resolving through the 3 layers.
func GetGeminiAPIKey() string {
	// Layer 1: Env vars
	if val := os.Getenv(cakdGeminiKeyEnv); val != "" {
		return val
	}
	if val := os.Getenv(geminiKeyEnv); val != "" {
		return val
	}

	// Layer 2: Credentials file
	creds, err := LoadCredentials()
	if err == nil && creds.Gemini.APIKey != "" {
		return creds.Gemini.APIKey
	}

	// Layer 3: Interactive CLI Prompt
	if isInteractive() {
		fmt.Println("⚠️  Gemini API Key not found in environment or local settings.")
		key, err := promptSecure("Enter Gemini API Key: ")
		if err == nil && key != "" {
			if promptYesNo("Save this API Key to ~/.cakd/credentials.yaml? [Y/n]: ") {
				creds, err = LoadCredentials()
				if err == nil {
					creds.Gemini.APIKey = key
					_ = SaveCredentials(creds)
					fmt.Println("✅ Saved Gemini API Key locally.")
				}
			}
			return key
		}
	}

	return ""
}
