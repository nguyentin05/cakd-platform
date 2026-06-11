package provider

import (
	"fmt"
	"os"

	"github.com/nguyentin05/cakd-platform/internal/provider/app"
	"github.com/nguyentin05/cakd-platform/internal/provider/app/springboot"
	"github.com/nguyentin05/cakd-platform/internal/provider/cd"
	"github.com/nguyentin05/cakd-platform/internal/provider/cd/argocd"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify/discord"
	"github.com/nguyentin05/cakd-platform/internal/provider/version_control"
	"github.com/nguyentin05/cakd-platform/internal/provider/version_control/github"
)

// ProviderRegistry acts as a factory for dynamically resolving provider implementations based on names.
type ProviderRegistry struct {
	AppFrameworks   map[string]func() app.AppFramework
	VersionControls map[string]func() version_control.VersionControl
	CDs             map[string]func() cd.CD
	Notifiers       map[string]func() notify.Notifier
}

var Providers = &ProviderRegistry{
	AppFrameworks: map[string]func() app.AppFramework{
		"java-spring-boot": func() app.AppFramework { return springboot.NewClient() },
	},
	VersionControls: map[string]func() version_control.VersionControl{
		"github": func() version_control.VersionControl { return github.NewClient() },
	},
	CDs: map[string]func() cd.CD{
		"argocd": func() cd.CD { return argocd.New() },
	},
	Notifiers: map[string]func() notify.Notifier{
		"discord": func() notify.Notifier {
			token := os.Getenv("DISCORD_BOT_TOKEN")
			guildID := os.Getenv("DISCORD_GUILD_ID")
			return discord.NewClient(token, guildID)
		},
	},
}

func GetAppProvider(name string) (app.AppFramework, error) {
	if factory, ok := Providers.AppFrameworks[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported app provider: %s", name)
}

func GetVersionControlProvider(name string) (version_control.VersionControl, error) {
	if factory, ok := Providers.VersionControls[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported version control provider: %s", name)
}

func GetCDProvider(name string) (cd.CD, error) {
	if factory, ok := Providers.CDs[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported CD provider: %s", name)
}

func GetNotifyProvider(name string) (notify.Notifier, error) {
	if factory, ok := Providers.Notifiers[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported notify provider: %s", name)
}
