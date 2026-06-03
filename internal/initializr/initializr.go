package initializr

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
)

const baseURL = "https://start.spring.io/starter.zip"

func Generate(cfg *config.PlatformConfig, outDir string) error {
	deps := []string{"web", "actuator"}

	if cfg.Spec.Features.Monitoring != nil && *cfg.Spec.Features.Monitoring {
		deps = append(deps, "prometheus")
	}

	if cfg.Spec.Dependencies.Database != nil {
		deps = append(deps, "data-jpa", "h2")
		switch cfg.Spec.Dependencies.Database.Type {
		case "postgresql":
			deps = append(deps, "postgresql")
		case "mysql":
			deps = append(deps, "mysql")
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
		cfg.Spec.Version,
		groupId,
		cfg.Metadata.Name,
		cfg.Metadata.Name,
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
