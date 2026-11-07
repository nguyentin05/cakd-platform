package alertdispatch_test

import (
	"context"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/statestore"
	"github.com/nguyentin05/cakd-platform/internal/usecase/alertdispatch"
)

type stubNotifier struct {
	batches [][]diagnostic.NotificationItem
}

func (s *stubNotifier) ProvisionWebhook(_ context.Context, _ string) (string, error) {
	return "https://discord.com/api/webhooks/stub", nil
}

func (s *stubNotifier) SendAlert(_ context.Context, _ string, items []diagnostic.NotificationItem) error {
	s.batches = append(s.batches, items)
	return nil
}

func (s *stubNotifier) SendMessage(_ context.Context, _, _ string) error {
	return nil
}

type stubAI struct{}

func (stubAI) Analyze(_ context.Context, _ string) (string, error) {
	return "Check DB pool exhaustion", nil
}

func TestDispatchDualPhaseAlertAndAIDiagnosis(t *testing.T) {
	store := statestore.NewStoreWithDir(t.TempDir())
	_ = store.SaveWebhook(context.Background(), "order-system", "https://discord.com/api/webhooks/order")

	notif := &stubNotifier{}
	svc := alertdispatch.NewService("https://discord.com/api/webhooks/default", notif, stubAI{}, store)

	err := svc.ProcessAlertPayload(context.Background(), diagnostic.AlertPayload{
		Status: "firing",
		Alerts: []diagnostic.Alert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "HighRestarts", "namespace": "order-system", "severity": "critical"},
				Annotations: map[string]string{"description": "Pod order-api restarted 5 times"},
			},
		},
	})
	if err != nil {
		t.Fatalf("ProcessAlertPayload failed: %v", err)
	}
	if len(notif.batches) != 2 {
		t.Fatalf("expected 2 notification batches (raw alert + AI diagnosis), got %d", len(notif.batches))
	}
}
