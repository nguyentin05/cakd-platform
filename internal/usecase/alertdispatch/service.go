package alertdispatch

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// Service processes Alertmanager webhook payloads, dispatches immediate alert notifications,
// and performs asynchronous or synchronous AI root-cause diagnosis for firing alerts.
type Service struct {
	defaultWebhookURL string
	notifier          ports.NotificationProvider
	analyzer          ports.AIAnalyzer
	stateStore        ports.ClusterStateStore
	aiMu              sync.Mutex
}

// NewService creates an alertdispatch Service.
func NewService(
	defaultWebhookURL string,
	notifier ports.NotificationProvider,
	analyzer ports.AIAnalyzer,
	stateStore ports.ClusterStateStore,
) *Service {
	return &Service{
		defaultWebhookURL: defaultWebhookURL,
		notifier:          notifier,
		analyzer:          analyzer,
		stateStore:        stateStore,
	}
}

type alertGroup struct {
	webhookURL  string
	items       []diagnostic.NotificationItem
	firingDescs []string
}

// ProcessAlertPayload routes alerts by namespace to their Discord webhook, dispatches raw alerts,
// and sends an AI diagnosis follow-up for firing alerts.
func (s *Service) ProcessAlertPayload(ctx context.Context, payload diagnostic.AlertPayload) error {
	var webhooks map[string]string
	if s.stateStore != nil {
		webhooks, _ = s.stateStore.LoadWebhooks()
	}

	groups := make(map[string]*alertGroup)
	for _, alert := range payload.Alerts {
		ns := alert.Labels["namespace"]
		targetURL := s.defaultWebhookURL
		if ns != "" && webhooks != nil {
			if u := webhooks[ns]; u != "" {
				targetURL = u
			}
		}
		if targetURL == "" {
			continue
		}

		grp, ok := groups[targetURL]
		if !ok {
			grp = &alertGroup{webhookURL: targetURL}
			groups[targetURL] = grp
		}

		desc := alert.Annotations["description"]
		if desc == "" {
			desc = "No description provided."
		}
		if alert.Status == "firing" {
			grp.firingDescs = append(grp.firingDescs, desc)
		}

		grp.items = append(grp.items, diagnostic.NotificationItem{
			Title:       fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Status), alert.Labels["alertname"]),
			Description: fmt.Sprintf("**Severity:** %s\n**Description:** %s", alert.Labels["severity"], desc),
			Severity:    alert.Status,
		})
	}

	for _, grp := range groups {
		if s.notifier != nil && len(grp.items) > 0 {
			if err := s.notifier.SendAlert(ctx, grp.webhookURL, grp.items); err != nil {
				return err
			}
		}
		if s.analyzer != nil && s.notifier != nil && len(grp.firingDescs) > 0 {
			s.runAIDiagnosis(ctx, grp.webhookURL, grp.firingDescs)
		}
	}
	return nil
}

func (s *Service) runAIDiagnosis(ctx context.Context, webhookURL string, descriptions []string) {
	s.aiMu.Lock()
	defer s.aiMu.Unlock()

	prompt := fmt.Sprintf(`You are an expert DevOps AI assistant.
The following Kubernetes alerts have just fired:
- %s

Provide a short, concise diagnosis of the likely root cause and 1-2 bullet points for immediate troubleshooting steps.
Keep the response under 150 words.`, strings.Join(descriptions, "\n- "))

	diagnosis, err := s.analyzer.Analyze(ctx, prompt)
	if err != nil || diagnosis == "" {
		return
	}

	_ = s.notifier.SendAlert(ctx, webhookURL, []diagnostic.NotificationItem{
		{
			Title:       "🤖 CAKD AI Diagnosis",
			Description: diagnosis,
			Severity:    "info",
		},
	})
}
