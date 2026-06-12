package springboot

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/registry"
	"github.com/nguyentin05/cakd-platform/internal/schema"
)

const baseURL = "https://start.spring.io/starter.zip"

// Client implements the [app.AppFramework] interface for Spring Boot.
// It interacts with start.spring.io (Spring Initializr API) to download zipped codebase templates.
type Client struct{}

// NewClient initializes and returns a new Spring Boot Initializr Client.
func NewClient() *Client {
	return &Client{}
}

// Scaffold constructs the Spring Initializr query parameters, queries start.spring.io,
// and extracts the downloaded archive into the designated outDir.
func (c *Client) Scaffold(cfg *schema.PlatformConfig, svc schema.Service, outDir string) error {
	deps := []string{"web", "actuator"}
	deps = append(deps, svc.Dependencies...)

	if dep, ok := registry.MonitoringDeps[cfg.Providers.Monitoring]; ok {
		deps = append(deps, dep)
	}

	for _, use := range svc.Uses {
		for _, b := range cfg.Backing {
			if b.Name == use {
				if springDeps, ok := registry.SpringBootDeps[b.Type]; ok {
					deps = append(deps, springDeps...)
				}
			}
		}
	}

	safeOwner := strings.ToLower(strings.ReplaceAll(cfg.Metadata.Owner, "-", ""))
	if safeOwner == "" {
		safeOwner = "example"
	}
	groupId := "com." + safeOwner

	projectType := svc.ProjectBuild
	if projectType == "" {
		projectType = "maven-project"
	}

	javaVersion := svc.LanguageVersion
	if javaVersion == "" {
		javaVersion = registry.Defaults.LanguageVersion["java-spring-boot"]
	}

	packaging := svc.Packaging
	if packaging == "" {
		packaging = "jar"
	}

	urlStr := fmt.Sprintf(
		"%s?type=%s&language=java&javaVersion=%s&packaging=%s&groupId=%s&artifactId=%s&name=%s&dependencies=%s",
		baseURL,
		projectType,
		javaVersion,
		packaging,
		groupId,
		svc.Name,
		svc.Name,
		strings.Join(deps, ","),
	)

	if svc.FrameworkVersion != "" {
		urlStr += "&bootVersion=" + svc.FrameworkVersion
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := httpClient.Get(urlStr) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to connect to start.spring.io: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("start.spring.io returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return extractZip(data, outDir)
}

// extractZip parses the zip byte payload and extracts all directories and files securely,
// preventing zip slip attacks, and applying correct permissions.
func extractZip(data []byte, outDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("failed to read zip archive: %w", err)
	}

	for _, file := range reader.File {
		targetPath := filepath.Join(outDir, file.Name) //nolint:gosec

		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(outDir)) {
			return fmt.Errorf("illegal file path in zip: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", targetPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("mkdir parent %s: %w", filepath.Dir(targetPath), err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", file.Name, err)
		}

		mode := file.FileInfo().Mode()
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create file %s: %w", targetPath, err)
		}

		if _, err := io.Copy(outFile, io.LimitReader(rc, 100*1024*1024)); err != nil {
			rc.Close()
			outFile.Close()
			return fmt.Errorf("extract file %s: %w", file.Name, err)
		}

		rc.Close()
		outFile.Close()

		if err := os.Chmod(targetPath, mode); err != nil {
			return fmt.Errorf("chmod file %s: %w", targetPath, err)
		}
	}

	os.Remove(filepath.Join(outDir, "src", "main", "resources", "application.properties"))

	return nil
}
