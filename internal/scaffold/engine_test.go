package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/config/defaults"
	"github.com/nguyentin05/cakd-platform/internal/scaffold"
)

func TestEngine_Generate(t *testing.T) {
	const dockerfilePath = "services/api-service/Dockerfile"

	tests := []struct {
		name         string
		cfg          *config.PlatformConfig
		expectedDocs []string // paths relative to output dir
		expectedYml  bool     // expect application.yml for service
		expectedProp bool     // expect application.properties for service
		verifyFile   string   // check contents of this file
		contains     string   // expected substring in verifyFile
	}{
		{
			name: "Gradle with Properties config, CI and CD enabled",
			cfg: &config.PlatformConfig{
				APIVersion: "platform.dev/v1alpha1",
				Kind:       "Project",
				Metadata: config.Metadata{
					Name:  "my-project",
					Owner: "tin",
				},
				Providers: config.Providers{
					VersionControl: "github",
					CI:             "github-actions",
					CD:             "argocd",
				},
				Services: []config.Service{
					{
						Name:               "api-service",
						Language:           "java-spring-boot",
						LanguageVersion:    "21",
						ProjectBuild:       "gradle-project",
						SpringConfigFormat: "properties",
					},
				},
			},
			expectedDocs: []string{
				".gitignore",
				".github/workflows/ci.yml",
				"deploy/application.yaml",
				dockerfilePath,
				"services/api-service/helm/Chart.yaml",
				"services/api-service/helm/values.yaml",
			},
			expectedProp: true,
			expectedYml:  false,
			verifyFile:   dockerfilePath,
			contains:     "./gradlew clean build", // Gradle build command
		},
		{
			name: "Maven with YAML config, CI enabled, CD disabled",
			cfg: &config.PlatformConfig{
				APIVersion: "platform.dev/v1alpha1",
				Kind:       "Project",
				Metadata: config.Metadata{
					Name:  "my-project",
					Owner: "tin",
				},
				Providers: config.Providers{
					VersionControl: "github",
					CI:             "github-actions",
				},
				Services: []config.Service{
					{
						Name:               "api-service",
						Language:           "java-spring-boot",
						LanguageVersion:    "21",
						ProjectBuild:       "maven-project",
						SpringConfigFormat: "yaml",
					},
				},
			},
			expectedDocs: []string{
				".gitignore",
				".github/workflows/ci.yml",
				dockerfilePath,
			},
			expectedProp: false,
			expectedYml:  true,
			verifyFile:   dockerfilePath,
			contains:     "mvn clean package", // Maven build command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			defaults.Apply(tt.cfg)
			engine := scaffold.New(tt.cfg)
			err := engine.Generate(tempDir)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			// 1. Verify expected files exist
			for _, fileRel := range tt.expectedDocs {
				path := filepath.Join(tempDir, fileRel)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Expected file %s to exist, but it does not", fileRel)
				}
			}

			// 2. Verify spring config format files
			propPath := filepath.Join(tempDir, "services", "api-service", "src", "main", "resources", "application.properties")
			ymlPath := filepath.Join(tempDir, "services", "api-service", "src", "main", "resources", "application.yml")

			if _, err := os.Stat(propPath); (err == nil) != tt.expectedProp {
				t.Errorf("Expected application.properties existence: %v, got: %v", tt.expectedProp, err == nil)
			}

			if _, err := os.Stat(ymlPath); (err == nil) != tt.expectedYml {
				t.Errorf("Expected application.yml existence: %v, got: %v", tt.expectedYml, err == nil)
			}

			// 3. Verify specific content in file
			if tt.verifyFile != "" && tt.contains != "" {
				contentBytes, err := os.ReadFile(filepath.Join(tempDir, tt.verifyFile))
				if err != nil {
					t.Fatalf("Failed to read file %s to verify: %v", tt.verifyFile, err)
				}
				content := string(contentBytes)
				if !strings.Contains(content, tt.contains) {
					t.Errorf("Expected file %s to contain %q, but it did not.\nFile content:\n%s", tt.verifyFile, tt.contains, content)
				}
			}
		})
	}
}
