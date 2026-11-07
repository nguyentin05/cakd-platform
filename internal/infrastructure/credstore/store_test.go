package credstore_test

import (
	"path/filepath"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/infrastructure/credstore"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

func TestStoreSaveGetDeleteAndEnvPrecedence(t *testing.T) {
	credFile := filepath.Join(t.TempDir(), "credentials.yaml")
	store := credstore.NewStoreWithPath(credFile)

	t.Setenv("CAKD_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	err := store.Save(ports.CredentialSet{
		GitHubToken:     "ghp_filetoken123456",
		DiscordBotToken: "disc_bottoken123456",
		DiscordGuildID:  "123456789",
		GeminiAPIKey:    "AIzaSyTestKey123456",
	})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := store.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.GitHubToken != "ghp_filetoken123456" {
		t.Fatalf("expected file token, got %q", got.GitHubToken)
	}

	t.Setenv("GITHUB_TOKEN", "ghp_envtoken999999")
	gotEnv, err := store.Get()
	if err != nil {
		t.Fatalf("Get with env failed: %v", err)
	}
	if gotEnv.GitHubToken != "ghp_envtoken999999" {
		t.Fatalf("expected env var to take precedence, got %q", gotEnv.GitHubToken)
	}

	if err := store.Delete("github"); err != nil {
		t.Fatalf("Delete(github) failed: %v", err)
	}
	t.Setenv("GITHUB_TOKEN", "")
	gotAfterDelete, err := store.Get()
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if gotAfterDelete.GitHubToken != "" {
		t.Fatalf("expected github token to be deleted, got %q", gotAfterDelete.GitHubToken)
	}
}
