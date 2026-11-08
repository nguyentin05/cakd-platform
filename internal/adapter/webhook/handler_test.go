package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/adapter/webhook"
	"github.com/nguyentin05/cakd-platform/internal/usecase/alertdispatch"
)

func TestWebhookHandlerAuthAndAccept(t *testing.T) {
	disp := alertdispatch.NewService("", nil, nil, nil)
	handler := webhook.NewHandler(disp, "my-secret")
	ts := httptest.NewServer(handler.Routes())
	defer ts.Close()

	// Healthz
	resHealth, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resHealth.Body.Close()
	if resHealth.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /healthz, got %d", resHealth.StatusCode)
	}

	// Unauthorized alert
	reqUnauth, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/alerts", strings.NewReader(`{"status":"firing"}`))
	reqUnauth.Header.Set("Content-Type", "application/json")
	resUnauth, err := http.DefaultClient.Do(reqUnauth)
	if err != nil {
		t.Fatal(err)
	}
	resUnauth.Body.Close()
	if resUnauth.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing secret, got %d", resUnauth.StatusCode)
	}

	// Authorized alert
	reqAuth, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/alerts", strings.NewReader(`{"status":"firing","alerts":[]}`))
	reqAuth.Header.Set("Content-Type", "application/json")
	reqAuth.Header.Set("Authorization", "Bearer my-secret")
	resAuth, err := http.DefaultClient.Do(reqAuth)
	if err != nil {
		t.Fatal(err)
	}
	resAuth.Body.Close()
	if resAuth.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 for valid secret, got %d", resAuth.StatusCode)
	}
}
