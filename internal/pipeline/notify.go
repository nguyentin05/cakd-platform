package pipeline

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/provider"
)

// NotifyStep provisions notification channels.
type NotifyStep struct{}

func (s *NotifyStep) Name() string { return "Provisioning Notification Channels" }

func (s *NotifyStep) ShouldRun(ctx *Context) bool {
	return ctx.Cfg.Providers.Notification != ""
}

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
