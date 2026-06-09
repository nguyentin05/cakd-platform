package factory

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/provider"
	"github.com/nguyentin05/cakd-platform/internal/provider/app/springboot"
	"github.com/nguyentin05/cakd-platform/internal/provider/iac/terraform"
	"github.com/nguyentin05/cakd-platform/internal/provider/llm/gemini"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify/discord"
	"github.com/nguyentin05/cakd-platform/internal/provider/vcs/github"
)

type Factory struct {
	cfg *config.PlatformConfig
}

func NewFactory(cfg *config.PlatformConfig) *Factory {
	return &Factory{cfg: cfg}
}

func (f *Factory) GetNotifier() (provider.Notifier, error) {
	providerName := f.cfg.Providers.Notification

	switch providerName {
	case "discord":
		token := os.Getenv("DISCORD_BOT_TOKEN")
		guildID := os.Getenv("DISCORD_GUILD_ID")
		return discord.NewClient(token, guildID), nil
	default:
		return nil, fmt.Errorf("unsupported notify provider: %s", providerName)
	}
}

func (f *Factory) GetLLM() (provider.LLM, error) {
	providerName := f.cfg.Providers.LLM

	switch providerName {
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		return gemini.NewClient(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", providerName)
	}
}

func (f *Factory) GetVCS() (provider.VCS, error) {
	providerName := f.cfg.Providers.VersionControl

	switch providerName {
	case "github":
		return github.NewClient(), nil
	default:
		return nil, fmt.Errorf("unsupported vcs provider: %s", providerName)
	}
}

func (f *Factory) GetAppFramework(svc config.Service) (provider.AppFramework, error) {
	providerName := svc.Language

	switch providerName {
	case "java-spring-boot":
		return springboot.NewClient(), nil
	default:
		return nil, fmt.Errorf("unsupported app framework: %s", providerName)
	}
}

func (f *Factory) GetIaC(workDir string) (provider.IaC, error) {
	providerName := "terraform"

	switch providerName {
	case "terraform":
		return terraform.New(f.cfg, workDir), nil
	default:
		return nil, fmt.Errorf("unsupported iac provider: %s", providerName)
	}
}
