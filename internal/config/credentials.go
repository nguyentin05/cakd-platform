package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Credentials represents the structured YAML configuration containing local authentication tokens
// and keys for various integrated providers (such as GitHub, Discord, and Gemini).
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

// GetCredentialsPath returns the absolute path to the local credentials file (~/.cakd/credentials.yaml).
// If a custom override path has been set, it returns the override path instead.
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

// LoadCredentials reads, parses, and returns the local credentials from (~/.cakd/credentials.yaml).
// If the credentials file does not exist, it returns an empty Credentials structure without error.
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

// SaveCredentials serializes the given credentials structure to YAML format and writes it
// to the local credentials file (~/.cakd/credentials.yaml) with read/write permissions restricted
// to the current user (0600). It creates the parent directory if it does not exist.
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

// isInteractive returns true if standard input is an interactive terminal.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptSecure prompts the user via terminal and reads a masked input. It is used for
// securely entering tokens and API keys without echoing the input characters to the screen.
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

// promptPlain prompts the user via terminal and reads a visible plain-text input line.
func promptPlain(prompt string) (string, error) {
	fmt.Print(prompt)
	var val string
	_, err := fmt.Scanln(&val)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

// promptYesNo displays a prompt and reads a boolean confirmation response. It interprets
// yes, y, or an empty input as confirmation (true), and any other inputs as denial (false).
func promptYesNo(prompt string) bool {
	fmt.Print(prompt)
	var response string
	_, _ = fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes"
}

// GetGithubToken retrieves the GitHub personal access token using a three-layer resolution order:
//
// 1. Checks environment variables (CAKD_GITHUB_TOKEN or GITHUB_TOKEN).
// 2. Looks up the token in the local credentials config file (~/.cakd/credentials.yaml).
// 3. Fallbacks to prompting the user interactively if stdin is a terminal, offering to save the token locally.
//
// Returns an empty string if the token cannot be resolved.
func GetGithubToken() string {
	if val := os.Getenv(cakdGithubTokenEnv); val != "" {
		return val
	}
	if val := os.Getenv(githubTokenEnv); val != "" {
		return val
	}

	creds, err := LoadCredentials()
	if err == nil && creds.Github.Token != "" {
		return creds.Github.Token
	}

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

// GetDiscordCredentials retrieves the Discord Bot Token and Guild ID using a three-layer resolution order:
//
// 1. Checks environment variables (CAKD_DISCORD_BOT_TOKEN/DISCORD_BOT_TOKEN and CAKD_DISCORD_GUILD_ID/DISCORD_GUILD_ID).
// 2. Looks up the values in the local credentials config file (~/.cakd/credentials.yaml).
// 3. Fallbacks to prompting the user interactively if stdin is a terminal, offering to save credentials locally.
//
// Returns an error if any of the required credentials cannot be resolved.
func GetDiscordCredentials() (token string, guildID string, err error) {
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

	if isInteractive() {
		fmt.Println("⚠️  Discord Bot Token or Guild ID not found in environment or local settings.")
		return promptDiscordCredentials(tok, gid)
	}

	return "", "", fmt.Errorf("discord credentials are not fully configured")
}

// promptDiscordCredentials displays interactive prompt steps to retrieve missing Discord credentials
// (Bot Token and/or Guild ID) and optionally saves them to the local configuration file.
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

// GetGeminiAPIKey retrieves the Gemini API Key using a three-layer resolution order:
//
// 1. Checks environment variables (CAKD_GEMINI_API_KEY or GEMINI_API_KEY).
// 2. Looks up the key in the local credentials config file (~/.cakd/credentials.yaml).
// 3. Fallbacks to prompting the user interactively if stdin is a terminal, offering to save the key locally.
//
// Returns an empty string if the key cannot be resolved.
func GetGeminiAPIKey() string {
	if val := os.Getenv(cakdGeminiKeyEnv); val != "" {
		return val
	}
	if val := os.Getenv(geminiKeyEnv); val != "" {
		return val
	}

	creds, err := LoadCredentials()
	if err == nil && creds.Gemini.APIKey != "" {
		return creds.Gemini.APIKey
	}

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
