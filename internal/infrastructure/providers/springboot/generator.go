package springboot

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
)

const defaultInitializrURL = "https://start.spring.io/starter.zip"

// Generator implements ports.FrameworkGenerator using the Spring Initializr API.
type Generator struct {
	baseURL    string
	httpClient *http.Client
}

// NewGenerator creates a Spring Boot Generator targeting start.spring.io.
func NewGenerator() *Generator {
	return &Generator{
		baseURL:    defaultInitializrURL,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

// NewGeneratorWithBaseURL creates a Spring Boot Generator with a custom base URL (for testing).
func NewGeneratorWithBaseURL(baseURL string) *Generator {
	return &Generator{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GenerateService downloads a starter archive from Spring Initializr and extracts it into destDir.
func (g *Generator) GenerateService(ctx context.Context, svc project.ServiceSpec, destDir string) error {
	deps := []string{"web", "actuator"}
	for _, d := range svc.Dependencies {
		if d != "web" && d != "actuator" {
			deps = append(deps, d)
		}
	}

	groupID := svc.GroupID
	if groupID == "" {
		groupID = "com.example"
	}
	artifactID := svc.ArtifactID
	if artifactID == "" {
		artifactID = svc.Name
	}
	buildType := svc.ProjectBuild
	if buildType == "" {
		buildType = "maven-project"
	}
	javaVersion := svc.Version
	if javaVersion == "" {
		javaVersion = catalog.Defaults.LanguageVersion[catalog.LangJavaSpringBoot]
	}
	packaging := svc.Packaging
	if packaging == "" {
		packaging = "jar"
	}

	q := url.Values{}
	q.Set("type", buildType)
	q.Set("language", "java")
	q.Set("javaVersion", javaVersion)
	q.Set("packaging", packaging)
	q.Set("groupId", groupID)
	q.Set("artifactId", artifactID)
	q.Set("name", svc.Name)
	q.Set("dependencies", strings.Join(deps, ","))
	if svc.SpringBootVersion != "" {
		q.Set("bootVersion", svc.SpringBootVersion)
	}

	reqURL := g.baseURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "create initializr request", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "query start.spring.io", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "read initializr response", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errcode.New(errcode.CodeExternal, fmt.Sprintf("start.spring.io returned status %d: %s", resp.StatusCode, string(body)))
	}

	return extractZip(body, destDir)
}

func extractZip(data []byte, outDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "open zip archive", err)
	}

	cleanOut := filepath.Clean(outDir)
	for _, file := range reader.File {
		//nolint:gosec
		targetPath := filepath.Join(outDir, file.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), cleanOut) {
			return errcode.New(errcode.CodeValidation, "illegal file path in zip: "+file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o750); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
			return err
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		mode := file.FileInfo().Mode()
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			_ = rc.Close()
			return err
		}
		if _, err := io.Copy(outFile, io.LimitReader(rc, 100*1024*1024)); err != nil {
			_ = rc.Close()
			_ = outFile.Close()
			return err
		}
		_ = rc.Close()
		_ = outFile.Close()
	}

	_ = os.Remove(filepath.Join(outDir, "src", "main", "resources", "application.properties"))
	return nil
}
