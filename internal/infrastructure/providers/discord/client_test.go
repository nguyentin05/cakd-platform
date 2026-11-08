package discord_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/discord"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

type mockCredStore struct {
	mu    sync.Mutex
	creds ports.CredentialSet
}

func (m *mockCredStore) Get() (ports.CredentialSet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.creds, nil
}

func (m *mockCredStore) Save(creds ports.CredentialSet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.creds = creds
	return nil
}

func (m *mockCredStore) Delete(_ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.creds = ports.CredentialSet{}
	return nil
}

func (m *mockCredStore) Status() (ports.CredentialStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return ports.CredentialStatus{
		DiscordBotToken: m.creds.DiscordBotToken,
		DiscordGuildID:  m.creds.DiscordGuildID,
	}, nil
}

func (m *mockCredStore) Path() string {
	return "test/credentials.yaml"
}

func TestClient_DynamicCredentialReload(t *testing.T) {
	store := &mockCredStore{}

	var mu sync.Mutex
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedAuthHeader = r.Header.Get("Authorization")
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/channels") && !strings.Contains(r.URL.Path, "/webhooks") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":   "chan-123",
				"name": "alerts-test-chan",
			})
			return
		}
		if strings.Contains(r.URL.Path, "/webhooks") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":    "wh-456",
				"token": "tok-789",
				"url":   "https://discord.com/api/webhooks/wh-456/tok-789",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := discord.NewClientWithStoreAndURL(store, server.URL)

	// Step 1: Without credentials, ProvisionWebhook must fail with CodeAuth
	_, err := client.ProvisionWebhook(context.Background(), "test-chan")
	if err == nil {
		t.Fatal("expected error when credentials missing, got nil")
	}
	if !errcode.IsCode(err, errcode.CodeAuth) {
		t.Fatalf("expected CodeAuth error, got: %v", err)
	}

	// Step 2: Set credentials dynamically into the store
	_ = store.Save(ports.CredentialSet{
		DiscordBotToken: "dynamic-token-xyz",
		DiscordGuildID:  "dynamic-guild-123",
	})

	// Step 3: Next ProvisionWebhook call must pick up dynamic credentials
	whURL, err := client.ProvisionWebhook(context.Background(), "test-chan")
	if err != nil {
		t.Fatalf("unexpected error after updating credentials: %v", err)
	}
	if whURL != "https://discord.com/api/webhooks/wh-456/tok-789" {
		t.Fatalf("expected webhook URL, got: %s", whURL)
	}

	mu.Lock()
	auth := receivedAuthHeader
	mu.Unlock()
	if auth != "Bot dynamic-token-xyz" {
		t.Fatalf("expected Authorization 'Bot dynamic-token-xyz', got: %s", auth)
	}
}

func TestClient_SendAlertAndMessage(t *testing.T) {
	var receivedPayloads []map[string]any
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		receivedPayloads = append(receivedPayloads, payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := discord.NewClient("static-token", "static-guild")

	// Empty alert or message should be no-op
	if err := client.SendAlert(context.Background(), server.URL, nil); err != nil {
		t.Fatalf("empty alert returned error: %v", err)
	}
	if err := client.SendMessage(context.Background(), server.URL, ""); err != nil {
		t.Fatalf("empty message returned error: %v", err)
	}

	// Non-empty alert
	items := []diagnostic.NotificationItem{
		{Title: "Critical Alert", Description: "Pod crashing", Severity: "critical"},
		{Title: "Warning Alert", Description: "High CPU", Severity: "warning"},
	}
	if err := client.SendAlert(context.Background(), server.URL, items); err != nil {
		t.Fatalf("SendAlert failed: %v", err)
	}

	// Non-empty message
	if err := client.SendMessage(context.Background(), server.URL, "Deployment finished"); err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	mu.Lock()
	count := len(receivedPayloads)
	mu.Unlock()

	if count != 2 {
		t.Fatalf("expected 2 received payloads, got %d", count)
	}
}
