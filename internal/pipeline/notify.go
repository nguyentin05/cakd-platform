package pipeline

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/provider"
)

// NotifyStep is an optional pipeline step responsible for provisioning notification channels
// (such as a Discord channel and webhook) for system alerts and deployment status.
type NotifyStep struct{}

// Name returns the step description.
func (s *NotifyStep) Name() string { return "Provisioning Notification Channels" }

// ShouldRun returns true if a notification provider is configured in platform.yaml.
func (s *NotifyStep) ShouldRun(ctx *Context) bool {
	return ctx.Cfg.Providers.Notification != ""
}

// Run creates the notification channel/webhook through the configured provider,
// then persists the routing webhook endpoint locally and in Kubernetes.
func (s *NotifyStep) Run(ctx *Context) error {
	notifier, err := provider.GetNotifyProvider(ctx.Cfg.Providers.Notification)
	if err != nil {
		return err
	}

	webhookURL, err := notifier.ProvisionChannel(ctx.Cfg.Metadata.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "   Warning: Failed to provision notification channel: %v\n", err)
		return nil
	}

	fmt.Printf("   Webhook created: %s\n", webhookURL)
	if err := config.SaveWebhook(ctx.Cfg.Metadata.Name, webhookURL); err != nil {
		fmt.Fprintf(os.Stderr, "   Warning: Failed to save webhook to local config: %v\n", err)
	} else {
		fmt.Printf("   Webhook routing rule saved for namespace: %s\n", ctx.Cfg.Metadata.Name)
	}

	return nil
}
