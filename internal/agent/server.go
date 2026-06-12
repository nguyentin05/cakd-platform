package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/provider/llm/gemini"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
)

// AgentServer runs the HTTP receiver for Alertmanager webhooks.
// It formats alerts, routes them to Discord channels, and triggers AI analysis.
type AgentServer struct {
	DefaultWebhookURL string
	notifier          notify.Notifier
	aiMutex           sync.Mutex
	geminiAPIKey      string
	geminiModel       string
	agentSecret       string
	llmClient         *gemini.Client
	webhooks          map[string]string
}

// NewAgentServer initializes and returns a new AgentServer instance.
func NewAgentServer(defaultWebhookURL string, notifier notify.Notifier, geminiAPIKey, geminiModel, agentSecret string, webhooks map[string]string) *AgentServer {
	var llmClient *gemini.Client
	if geminiAPIKey != "" {
		llmClient = gemini.NewClient(geminiAPIKey, geminiModel)
	}
	return &AgentServer{
		DefaultWebhookURL: defaultWebhookURL,
		notifier:          notifier,
		geminiAPIKey:      geminiAPIKey,
		geminiModel:       geminiModel,
		agentSecret:       agentSecret,
		llmClient:         llmClient,
		webhooks:          webhooks,
	}
}

// HandleAlert accepts Alertmanager JSON webhook POST requests, sends a quick
// acceptance response, and processes the alerts asynchronously.
func (s *AgentServer) HandleAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	// Validate shared secret if configured
	if s.agentSecret != "" {
		authHeader := r.Header.Get("Authorization")
		querySecret := r.URL.Query().Get("secret")
		authorized := false
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token == s.agentSecret {
					authorized = true
				}
			}
		}
		if querySecret == s.agentSecret {
			authorized = true
		}
		if !authorized {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Limit body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024)

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "accepted"}`))

	go s.processAlerts(payload)
}

// processAlerts groups incoming alerts by project/namespace webhook URLs, formats their severity
// and description details, dispatches them via the notifier, and queues them for AI analysis.
func (s *AgentServer) processAlerts(payload AlertmanagerPayload) {
	type AlertGroup struct {
		WebhookURL  string
		Items       []notify.AlertItem
		FiringDescs []string
	}
	groups := make(map[string]*AlertGroup)

	for _, alert := range payload.Alerts {
		namespace := alert.Labels["namespace"]
		targetURL := s.DefaultWebhookURL
		if namespace != "" && s.webhooks != nil {
			if url, ok := s.webhooks[namespace]; ok && url != "" {
				targetURL = url
			}
		}

		if _, exists := groups[targetURL]; !exists {
			groups[targetURL] = &AlertGroup{WebhookURL: targetURL}
		}

		if alert.Status == "firing" {
			groups[targetURL].FiringDescs = append(groups[targetURL].FiringDescs, alert.Annotations["description"])
		}

		alertName := alert.Labels["alertname"]
		description := alert.Annotations["description"]
		if description == "" {
			description = "No description provided."
		}

		desc := fmt.Sprintf("**Severity:** %s\n**Description:** %s", alert.Labels["severity"], description)
		item := notify.AlertItem{
			Title:       fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Status), alertName),
			Description: desc,
			Severity:    alert.Status,
		}
		groups[targetURL].Items = append(groups[targetURL].Items, item)
	}

	for _, group := range groups {
		if len(group.Items) > 0 {
			msg := notify.AlertPayload{Items: group.Items}
			if err := s.notifier.SendAlert(group.WebhookURL, msg); err != nil {
				fmt.Printf("Failed to send raw alert via notifier: %v\n", err)
			}
		}

		if len(group.FiringDescs) > 0 {
			go s.runAIAnalysis(group.WebhookURL, group.FiringDescs)
		}
	}
}

// runAIAnalysis uses the Gemini LLM client to analyze firing alert descriptions and
// posts troubleshooting suggestions back to the project's alert channel.
func (s *AgentServer) runAIAnalysis(targetWebhookURL string, descriptions []string) {
	if s.llmClient == nil {
		fmt.Println("Gemini client is not initialized. Skipping AI analysis.")
		return
	}

	time.Sleep(3 * time.Second)

	s.aiMutex.Lock()
	defer s.aiMutex.Unlock()

	fmt.Println("Running AI analysis for fired alerts...")

	alertContext := strings.Join(descriptions, "\n- ")
	prompt := fmt.Sprintf(`You are an expert DevOps AI assistant. 
The following Kubernetes alerts have just fired:
- %s

Provide a very short, concise diagnosis of the likely root cause and 1-2 bullet points for immediate troubleshooting steps. 
Keep the response under 150 words. Do not use markdown codeblocks that wrap the entire response.`, alertContext)

	diagnosis, err := s.llmClient.Analyze(prompt)
	if err != nil {
		fmt.Printf("Failed to generate AI diagnosis: %v\n", err)
		return
	}

	aiMsg := notify.AlertPayload{
		Items: []notify.AlertItem{
			{
				Title:       "🤖 CAKD AI Diagnosis",
				Description: diagnosis,
				Severity:    "info",
			},
		},
	}

	if err := s.notifier.SendAlert(targetWebhookURL, aiMsg); err != nil {
		fmt.Printf("Failed to send AI diagnosis via notifier: %v\n", err)
	} else {
		fmt.Println("AI diagnosis sent successfully.")
	}
}
