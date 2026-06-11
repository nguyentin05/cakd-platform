package app

import "github.com/nguyentin05/cakd-platform/internal/config"

// AppFramework generates a base project for a specific language/framework.
type AppFramework interface {
	Scaffold(cfg *config.PlatformConfig, svc config.Service, outDir string) error
}
