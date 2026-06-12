package app

import "github.com/nguyentin05/cakd-platform/internal/schema"

// AppFramework defines the interface for generators that bootstrap a service's base code skeleton.
// Implementations of this interface generate boilerplate configurations and source file layouts
// for specific programming languages or web frameworks (e.g. Java Spring Boot).
type AppFramework interface {
	// Scaffold generates the boilerplate skeleton code and directories for a service inside the target outDir.
	Scaffold(cfg *schema.PlatformConfig, svc schema.Service, outDir string) error
}
