package providers_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/discord"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/gemini"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/springboot"
)

func TestSpringBootGeneratorExtractsZip(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("pom.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte("<project></project>"))
	_ = zw.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipBuf.Bytes())
	}))
	defer srv.Close()

	gen := springboot.NewGeneratorWithBaseURL(srv.URL)
	destDir := t.TempDir()
	err = gen.GenerateService(context.Background(), project.ServiceSpec{
		Name:         "order-api",
		Language:     "java-spring-boot",
		Version:      "21",
		Dependencies: []string{"web", "actuator"},
	}, destDir)
	if err != nil {
		t.Fatalf("GenerateService failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "pom.xml")); err != nil {
		t.Fatalf("expected pom.xml to be extracted: %v", err)
	}
}

func TestDiscordClientSendAlertAndMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := discord.NewClient("bot-token", "guild-123")
	ctx := context.Background()
	err := client.SendAlert(ctx, srv.URL, []diagnostic.NotificationItem{
		{Title: "[FIRING] PodCrashLooping", Description: "Pod restarted", Severity: "firing"},
	})
	if err != nil {
		t.Fatalf("SendAlert failed: %v", err)
	}
	if err := client.SendMessage(ctx, srv.URL, "Diagnostic summary"); err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
}

func TestGeminiClientAnalyze(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]string{
							{"text": "Root cause: DB connection refused"},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	client := gemini.NewClientWithEndpoint("test-key", "gemini-2.5-flash", srv.URL)
	ans, err := client.Analyze(context.Background(), "diagnose this")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if ans != "Root cause: DB connection refused" {
		t.Fatalf("unexpected analysis: %q", ans)
	}
}
