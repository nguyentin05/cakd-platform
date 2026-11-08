package alertdispatch_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/statestore"
	"github.com/nguyentin05/cakd-platform/internal/usecase/alertdispatch"
)

type stubNotifier struct {
	mu      sync.Mutex
	batches [][]diagnostic.NotificationItem
}

func (s *stubNotifier) ProvisionWebhook(_ context.Context, _ string) (string, error) {
	return "https://discord.com/api/webhooks/stub", nil
}

func (s *stubNotifier) SendAlert(_ context.Context, _ string, items []diagnostic.NotificationItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batches = append(s.batches, items)
	return nil
}

func (s *stubNotifier) SendMessage(_ context.Context, _, _ string) error {
	return nil
}

func (s *stubNotifier) BatchCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.batches)
}

type stubAI struct {
	concurrent int64
	maxSeen    int64
}

func (s *stubAI) Analyze(_ context.Context, _ string) (string, error) {
	cur := atomic.AddInt64(&s.concurrent, 1)
	defer atomic.AddInt64(&s.concurrent, -1)

	for {
		old := atomic.LoadInt64(&s.maxSeen)
		if cur <= old || atomic.CompareAndSwapInt64(&s.maxSeen, old, cur) {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	return "Check DB pool exhaustion", nil
}

func TestDispatchDualPhaseAlertAndAIDiagnosis(t *testing.T) {
	store := statestore.NewStoreWithDir(t.TempDir())
	_ = store.SaveWebhook(context.Background(), "order-system", "https://discord.com/api/webhooks/order")

	notif := &stubNotifier{}
	svc := alertdispatch.NewService("https://discord.com/api/webhooks/default", notif, &stubAI{}, store)

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
	if notif.BatchCount() != 2 {
		t.Fatalf("expected 2 notification batches (raw alert + AI diagnosis), got %d", notif.BatchCount())
	}
}

func TestConcurrentAlertStorm(t *testing.T) {
	store := statestore.NewStoreWithDir(t.TempDir())
	_ = store.SaveWebhook(context.Background(), "order-system", "https://discord.com/api/webhooks/order")

	notif := &stubNotifier{}
	ai := &stubAI{}
	svc := alertdispatch.NewService("https://discord.com/api/webhooks/default", notif, ai, store)

	const totalAlerts = 10
	var wg sync.WaitGroup
	errCh := make(chan error, totalAlerts)

	for i := 0; i < totalAlerts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := svc.ProcessAlertPayload(context.Background(), diagnostic.AlertPayload{
				Status: "firing",
				Alerts: []diagnostic.Alert{
					{
						Status:      "firing",
						Labels:      map[string]string{"alertname": "CrashLoopBackOff", "namespace": "order-system", "severity": "critical"},
						Annotations: map[string]string{"description": "Pod crashed repeatedly"},
					},
				},
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("concurrent ProcessAlertPayload failed: %v", err)
	}

	// All 10 raw alerts + 10 diagnoses must have completed without error
	if count := notif.BatchCount(); count != 20 {
		t.Fatalf("expected 20 batches (10 raw + 10 AI), got %d", count)
	}

	// Verify maximum concurrent AI calls never exceeded 3
	if maxConcurrent := atomic.LoadInt64(&ai.maxSeen); maxConcurrent > 3 {
		t.Fatalf("semaphore leaked: maximum concurrent AI executions was %d, expected <= 3", maxConcurrent)
	}
}
