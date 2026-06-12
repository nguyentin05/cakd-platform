package provider

import (
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/provider/app"
	"github.com/nguyentin05/cakd-platform/internal/provider/app/springboot"
	"github.com/nguyentin05/cakd-platform/internal/provider/cd"
	"github.com/nguyentin05/cakd-platform/internal/provider/cd/argocd"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify/discord"
	"github.com/nguyentin05/cakd-platform/internal/provider/version_control"
	"github.com/nguyentin05/cakd-platform/internal/provider/version_control/github"
)

// ProviderRegistry acts as a central factory registry for mapping provider names
// to dynamic initialization factories for application generators, VCS, CD, and notification systems.
type ProviderRegistry struct {
	AppFrameworks   map[string]func() app.AppFramework
	VersionControls map[string]func() version_control.VersionControl
	CDs             map[string]func() cd.CD
	Notifiers       map[string]func() notify.Notifier
}

// Providers is the global instance of the ProviderRegistry pre-populated with all supported integrations.
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
			token, guildID, err := config.GetDiscordCredentials()
			if err != nil {
				fmt.Printf("Warning: Discord credentials not configured: %v. Discord notifications will not be functional.\n", err)
			}
			return discord.NewClient(token, guildID)
		},
	},
}

// GetAppProvider resolves and returns an AppFramework generator implementation by name.
func GetAppProvider(name string) (app.AppFramework, error) {
	if factory, ok := Providers.AppFrameworks[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported app provider: %s", name)
}

// GetVersionControlProvider resolves and returns a VersionControl integration client by name.
func GetVersionControlProvider(name string) (version_control.VersionControl, error) {
	if factory, ok := Providers.VersionControls[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported version control provider: %s", name)
}

// GetCDProvider resolves and returns a continuous deployment integration client by name.
func GetCDProvider(name string) (cd.CD, error) {
	if factory, ok := Providers.CDs[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported CD provider: %s", name)
}

// GetNotifyProvider resolves and returns a notification integration client by name.
func GetNotifyProvider(name string) (notify.Notifier, error) {
	if factory, ok := Providers.Notifiers[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported notify provider: %s", name)
}
