package alertdispatch

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	aiSem             chan struct{}
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
		aiSem:             make(chan struct{}, 3),
	}
}

type alertGroup struct {
	webhookURL  string
	items       []diagnostic.NotificationItem
	firingDescs []string
}

// ProcessAlertPayload routes alerts by namespace to their Discord webhook, dispatches raw alerts,
// and sends an AI diagnosis follow-up for firing alerts.
//
//nolint:gocyclo
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
			unique := deduplicate(grp.firingDescs)
			if len(unique) > 0 {
				s.runAIDiagnosis(ctx, grp.webhookURL, unique)
			}
		}
	}
	return nil
}

func (s *Service) runAIDiagnosis(ctx context.Context, webhookURL string, descriptions []string) {
	if len(descriptions) == 0 {
		return
	}
	sem := s.aiSem
	if sem == nil {
		sem = make(chan struct{}, 3)
	}

	aiCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-aiCtx.Done():
		return
	}

	prompt := fmt.Sprintf(`You are an expert DevOps AI assistant.
The following Kubernetes alerts have just fired:
- %s

Provide a short, concise diagnosis of the likely root cause and 1-2 bullet points for immediate troubleshooting steps.
Keep the response under 150 words.`, strings.Join(descriptions, "\n- "))

	diagnosis, err := s.analyzer.Analyze(aiCtx, prompt)
	if err != nil || diagnosis == "" {
		return
	}

	_ = s.notifier.SendAlert(aiCtx, webhookURL, []diagnostic.NotificationItem{
		{
			Title:       "🤖 CAKD AI Diagnosis",
			Description: diagnosis,
			Severity:    "info",
		},
	})
}

func deduplicate(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}
	return result
}
