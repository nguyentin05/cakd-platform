package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials for providers (GitHub, Discord, Gemini)",
}

var loginCmd = &cobra.Command{
	Use:   "login [provider]",
	Short: "Log in to a provider (github, discord, gemini)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := strings.ToLower(args[0])
		creds, err := config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		switch provider {
		case "github":
			fmt.Print("Enter GitHub Personal Access Token (PAT): ")
			byteToken, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			token := strings.TrimSpace(string(byteToken))
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}
			creds.Github.Token = token
			fmt.Println("GitHub credentials prepared.")

		case "discord":
			fmt.Print("Enter Discord Bot Token: ")
			byteToken, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			token := strings.TrimSpace(string(byteToken))
			if token == "" {
				return fmt.Errorf("bot token cannot be empty")
			}

			fmt.Print("Enter Discord Guild ID (Server ID): ")
			guildIDInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read guild ID: %w", err)
			}
			guildID := strings.TrimSpace(guildIDInput)
			if guildID == "" {
				return fmt.Errorf("guild ID cannot be empty")
			}

			creds.Discord.Token = token
			creds.Discord.GuildID = guildID
			fmt.Println("Discord credentials prepared.")

		case "gemini":
			fmt.Print("Enter Gemini API Key: ")
			byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read api key: %w", err)
			}
			key := strings.TrimSpace(string(byteKey))
			if key == "" {
				return fmt.Errorf("api key cannot be empty")
			}
			creds.Gemini.APIKey = key
			fmt.Println("Gemini credentials prepared.")

		default:
			return fmt.Errorf("unknown provider %q. Supported: github, discord, gemini", provider)
		}

		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}
		fmt.Printf("Successfully logged in and saved credentials to config for: %s\n", provider)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status for all providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		path, _ := config.GetCredentialsPath()
		fmt.Printf("Credentials file path: %s\n\n", path)

		fmt.Println("Provider Authentication Status:")
		fmt.Printf("  GitHub:  %s\n", maskSecret(creds.Github.Token))
		fmt.Printf("  Discord: %s (Guild ID: %s)\n", maskSecret(creds.Discord.Token), maskGuildID(creds.Discord.GuildID))
		fmt.Printf("  Gemini:  %s\n", maskSecret(creds.Gemini.APIKey))

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout [provider]",
	Short: "Log out from a provider (github, discord, gemini)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := strings.ToLower(args[0])
		creds, err := config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("failed to load credentials: %w", err)
		}

		switch provider {
		case "github":
			creds.Github.Token = ""
		case "discord":
			creds.Discord.Token = ""
			creds.Discord.GuildID = ""
		case "gemini":
			creds.Gemini.APIKey = ""
		default:
			return fmt.Errorf("unknown provider %q. Supported: github, discord, gemini", provider)
		}

		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}
		fmt.Printf("Successfully logged out from: %s\n", provider)
		return nil
	},
}

func maskSecret(secret string) string {
	if secret == "" {
		return "❌ Not Authenticated"
	}
	if len(secret) <= 8 {
		return "✅ Authenticated (Hidden)"
	}
	return fmt.Sprintf("✅ Authenticated (ends in ...%s)", secret[len(secret)-4:])
}

func maskGuildID(id string) string {
	if id == "" {
		return "not set"
	}
	return id
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)
}
