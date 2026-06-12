package agent_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/agent"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
)

type MockNotifier struct {
	SendAlertFunc func(webhookURL string, payload notify.AlertPayload) error
	Sent          bool
}

func (m *MockNotifier) ProvisionChannel(projectName string) (webhookURL string, err error) {
	return "http://mock-webhook.discord/" + projectName, nil
}

func (m *MockNotifier) SendAlert(webhookURL string, payload notify.AlertPayload) error {
	m.Sent = true
	if m.SendAlertFunc != nil {
		return m.SendAlertFunc(webhookURL, payload)
	}
	return nil
}

func TestAgentServer_HandleAlert(t *testing.T) {
	mockNotifier := &MockNotifier{}
	server := agent.NewAgentServer("http://default-webhook.discord", mockNotifier, "", "", "", nil)

	t.Run("Valid POST payload", func(t *testing.T) {
		payload := `{
			"receiver": "webhook",
			"status": "firing",
			"alerts": [
				{
					"status": "firing",
					"labels": {
						"alertname": "DeploymentCPUHigh",
						"severity": "critical",
						"namespace": "my-app"
					},
					"annotations": {
						"description": "CPU usage is above 90%"
					}
				}
			]
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.HandleAlert(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		expectedBody := `{"status": "accepted"}`
		if !strings.Contains(w.Body.String(), expectedBody) {
			t.Errorf("Expected body to contain %q, got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("Invalid HTTP Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
		w := httptest.NewRecorder()

		server.HandleAlert(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("Malformed JSON payload", func(t *testing.T) {
		payload := `{"receiver": "webhook", "status":`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.HandleAlert(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Missing or Incorrect Content-Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		server.HandleAlert(w, req)

		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status code %d, got %d", http.StatusUnsupportedMediaType, w.Code)
		}
	})

	t.Run("Payload exceeds 1MB limit", func(t *testing.T) {
		largePayload := `{"receiver": "webhook", "status": "firing", "alerts": [`
		largePayload += strings.Repeat(`{"status":"firing","labels":{"alertname":"CPUHigh","severity":"critical","namespace":"app"},"annotations":{"description":"CPU usage is above 90%"}},`, 8000)
		largePayload = strings.TrimSuffix(largePayload, ",") + `]}`

		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(largePayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.HandleAlert(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Authentication - Unauthorized", func(t *testing.T) {
		secureServer := agent.NewAgentServer("http://default-webhook.discord", mockNotifier, "", "", "supersecret", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		secureServer.HandleAlert(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("Authentication - Authorized via Header", func(t *testing.T) {
		secureServer := agent.NewAgentServer("http://default-webhook.discord", mockNotifier, "", "", "supersecret", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(`{"alerts":[]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer supersecret")
		w := httptest.NewRecorder()

		secureServer.HandleAlert(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Authentication - Authorized via Query Param", func(t *testing.T) {
		secureServer := agent.NewAgentServer("http://default-webhook.discord", mockNotifier, "", "", "supersecret", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts?secret=supersecret", bytes.NewBufferString(`{"alerts":[]}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		secureServer.HandleAlert(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})
}
