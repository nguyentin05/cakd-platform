package webhook

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/usecase/alertdispatch"
)

// Handler serves the inbound Alertmanager webhook HTTP endpoint for cakd-agent.
type Handler struct {
	dispatcher *alertdispatch.Service
	secret     string
}

// NewHandler creates an inbound Alertmanager webhook HTTP Handler.
func NewHandler(dispatcher *alertdispatch.Service, secret string) *Handler {
	return &Handler{
		dispatcher: dispatcher,
		secret:     secret,
	}
}

// Routes returns an http.Handler serving POST /api/v1/alerts and GET /healthz.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealthz)
	mux.HandleFunc("/api/v1/alerts", h.HandleAlert)
	return mux
}

func (h *Handler) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// HandleAlert validates and accepts Alertmanager webhook POST requests, then dispatches them.
func (h *Handler) HandleAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	if h.secret != "" && !h.isAuthorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024)
	var payload diagnostic.AlertPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))

	if h.dispatcher != nil {
		go func(p diagnostic.AlertPayload) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_ = h.dispatcher.ProcessAlertPayload(ctx, p)
		}(payload)
	}
}

func (h *Handler) isAuthorized(r *http.Request) bool {
	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(token), []byte(h.secret)) == 1 {
			return true
		}
	}
	if q := r.URL.Query().Get("secret"); q != "" {
		if subtle.ConstantTimeCompare([]byte(q), []byte(h.secret)) == 1 {
			return true
		}
	}
	return false
}
