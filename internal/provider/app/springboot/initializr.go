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

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/registry"
)

const baseURL = "https://start.spring.io/starter.zip"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Scaffold(cfg *config.PlatformConfig, svc config.Service, outDir string) error {
	deps := []string{"web", "actuator"}

	// Monitoring dependency lookup from registry
	if dep, ok := registry.MonitoringDeps[cfg.Providers.Monitoring]; ok {
		deps = append(deps, dep)
	}

	// Backing resource dependency lookup from registry
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

	url := fmt.Sprintf(
		"%s?type=maven-project&language=java&javaVersion=%s&groupId=%s&artifactId=%s&name=%s&dependencies=%s",
		baseURL,
		svc.Version,
		groupId,
		svc.Name,
		svc.Name,
		strings.Join(deps, ","),
	)

	resp, err := http.Get(url) //nolint:gosec
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

		outFile, err := os.Create(targetPath)
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
	}

	return nil
}
