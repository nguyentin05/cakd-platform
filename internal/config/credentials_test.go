package config

import (
	"os"
	"path/filepath"
	"testing"
)

const testFileStoredToken = "file-stored-token"

func TestGetGithubToken_Layer1_EnvPrecedence(t *testing.T) {
	// Setup env variables
	t.Setenv(cakdGithubTokenEnv, "cakd-custom-env-token")
	t.Setenv(githubTokenEnv, "regular-env-token")

	// Set override path to a temp dir so it doesn't touch local home configs
	tempDir := t.TempDir()
	credentialsPathOverride = filepath.Join(tempDir, "credentials.yaml")
	defer func() { credentialsPathOverride = "" }()

	// Write mock file to Layer 2 to ensure env overrides file config
	creds := &Credentials{}
	creds.Github.Token = testFileStoredToken
	err := SaveCredentials(creds)
	if err != nil {
		t.Fatalf("Failed to save credentials: %v", err)
	}

	// Resolve and assert
	token := GetGithubToken()
	if token != "cakd-custom-env-token" {
		t.Errorf("Expected token to be 'cakd-custom-env-token', got: %s", token)
	}

	// Test fallback to regular GITHUB_TOKEN
	t.Setenv(cakdGithubTokenEnv, "")
	token = GetGithubToken()
	if token != "regular-env-token" {
		t.Errorf("Expected token to fallback to 'regular-env-token', got: %s", token)
	}
}

func TestGetGithubToken_Layer2_FileStore(t *testing.T) {
	// Unset env vars
	t.Setenv(cakdGithubTokenEnv, "")
	t.Setenv(githubTokenEnv, "")

	// Redirect path to temp dir
	tempDir := t.TempDir()
	credentialsPathOverride = filepath.Join(tempDir, "credentials.yaml")
	defer func() { credentialsPathOverride = "" }()

	// Save token in config file
	creds := &Credentials{}
	creds.Github.Token = testFileStoredToken
	err := SaveCredentials(creds)
	if err != nil {
		t.Fatalf("Failed to save credentials: %v", err)
	}

	// Resolve and assert
	token := GetGithubToken()
	if token != testFileStoredToken {
		t.Errorf("Expected token to be '%s', got: %s", testFileStoredToken, token)
	}
}

func TestGetDiscordCredentials_Precedence(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPathOverride = filepath.Join(tempDir, "credentials.yaml")
	defer func() { credentialsPathOverride = "" }()

	// Test Layer 1 (Env variables)
	t.Setenv(cakdDiscordTokenEnv, "cakd-bot-token")
	t.Setenv(cakdDiscordGuildEnv, "cakd-guild-id")
	t.Setenv(discordTokenEnv, "normal-bot-token")
	t.Setenv(discordGuildEnv, "normal-guild-id")

	token, guildID, err := GetDiscordCredentials()
	if err != nil {
		t.Fatalf("Failed to get Discord credentials: %v", err)
	}
	if token != "cakd-bot-token" || guildID != "cakd-guild-id" {
		t.Errorf("Expected CAKD env variables, got token: %s, guildID: %s", token, guildID)
	}

	// Fallback to normal env variables
	t.Setenv(cakdDiscordTokenEnv, "")
	t.Setenv(cakdDiscordGuildEnv, "")
	token, guildID, err = GetDiscordCredentials()
	if err != nil {
		t.Fatalf("Failed to get Discord credentials: %v", err)
	}
	if token != "normal-bot-token" || guildID != "normal-guild-id" {
		t.Errorf("Expected fallback normal env variables, got token: %s, guildID: %s", token, guildID)
	}

	// Test Layer 2 (File storage)
	t.Setenv(discordTokenEnv, "")
	t.Setenv(discordGuildEnv, "")

	creds := &Credentials{}
	creds.Discord.Token = "file-bot-token"
	creds.Discord.GuildID = "file-guild-id"
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("Failed to save credentials: %v", err)
	}

	token, guildID, err = GetDiscordCredentials()
	if err != nil {
		t.Fatalf("Failed to get Discord credentials: %v", err)
	}
	if token != "file-bot-token" || guildID != "file-guild-id" {
		t.Errorf("Expected file credentials, got token: %s, guildID: %s", token, guildID)
	}
}

func TestGetGeminiAPIKey_Precedence(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPathOverride = filepath.Join(tempDir, "credentials.yaml")
	defer func() { credentialsPathOverride = "" }()

	// Test Layer 1
	t.Setenv(cakdGeminiKeyEnv, "cakd-gemini-key")
	t.Setenv(geminiKeyEnv, "normal-gemini-key")

	key := GetGeminiAPIKey()
	if key != "cakd-gemini-key" {
		t.Errorf("Expected CAKD gemini API key, got: %s", key)
	}

	// Fallback to normal
	t.Setenv(cakdGeminiKeyEnv, "")
	key = GetGeminiAPIKey()
	if key != "normal-gemini-key" {
		t.Errorf("Expected normal gemini API key, got: %s", key)
	}

	// Test Layer 2
	t.Setenv(geminiKeyEnv, "")
	creds := &Credentials{}
	creds.Gemini.APIKey = "file-gemini-key"
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("Failed to save credentials: %v", err)
	}

	key = GetGeminiAPIKey()
	if key != "file-gemini-key" {
		t.Errorf("Expected file gemini API key, got: %s", key)
	}
}

func TestCredentials_LoadSave(t *testing.T) {
	tempDir := t.TempDir()
	credentialsPathOverride = filepath.Join(tempDir, "credentials.yaml")
	defer func() { credentialsPathOverride = "" }()

	// Ensure loading from non-existent file returns empty credentials without error
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials on non-existent file failed: %v", err)
	}
	if creds.Github.Token != "" || creds.Discord.Token != "" || creds.Gemini.APIKey != "" {
		t.Error("Expected empty credentials on non-existent file")
	}

	// Save credentials
	creds.Github.Token = "gh-123"
	creds.Discord.Token = "d-123"
	creds.Discord.GuildID = "dg-123"
	creds.Gemini.APIKey = "gem-123"
	err = SaveCredentials(creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	// Verify file was written with correct permissions
	info, err := os.Stat(credentialsPathOverride)
	if err != nil {
		t.Fatalf("Failed to stat credentials file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Expected file permissions to be 0600, got: %v", perm)
	}

	// Load and verify
	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}
	if loaded.Github.Token != "gh-123" || loaded.Discord.Token != "d-123" ||
		loaded.Discord.GuildID != "dg-123" || loaded.Gemini.APIKey != "gem-123" {
		t.Errorf("Loaded credentials did not match saved: %+v", loaded)
	}
}
