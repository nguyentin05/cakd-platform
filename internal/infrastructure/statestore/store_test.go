package statestore_test

import (
	"context"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/infrastructure/statestore"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

func TestClusterAndWebhookStatePersistence(t *testing.T) {
	baseDir := t.TempDir()
	store := statestore.NewStoreWithDir(baseDir)

	state := ports.ClusterState{
		Provider:       "aws-ec2",
		Host:           "54.10.20.30",
		Domain:         "54.10.20.30.nip.io",
		KubeconfigPath: "/tmp/k3s.yaml",
	}
	if err := store.SaveClusterState(state); err != nil {
		t.Fatalf("SaveClusterState failed: %v", err)
	}

	loaded, err := store.LoadClusterState()
	if err != nil {
		t.Fatalf("LoadClusterState failed: %v", err)
	}
	if loaded.Domain != "54.10.20.30.nip.io" || loaded.Host != "54.10.20.30" {
		t.Fatalf("unexpected cluster state: %+v", loaded)
	}

	ctx := context.Background()
	if err := store.SaveWebhook(ctx, "order-system", "https://discord.com/api/webhooks/123"); err != nil {
		t.Fatalf("SaveWebhook failed: %v", err)
	}

	hooks, err := store.LoadWebhooks()
	if err != nil {
		t.Fatalf("LoadWebhooks failed: %v", err)
	}
	if hooks["order-system"] != "https://discord.com/api/webhooks/123" {
		t.Fatalf("expected webhook for order-system, got %+v", hooks)
	}
}
