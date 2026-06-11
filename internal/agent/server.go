package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/provider/llm/gemini"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
)

type AgentServer struct {
	DefaultWebhookURL string
	notifier          notify.Notifier
	aiMutex           sync.Mutex
}

func NewAgentServer(defaultWebhookURL string, notifier notify.Notifier) *AgentServer {
	return &AgentServer{
		DefaultWebhookURL: defaultWebhookURL,
		notifier:          notifier,
	}
}

func (s *AgentServer) HandleAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload AlertmanagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "accepted"}`))

	go s.processAlerts(payload)
}

func (s *AgentServer) processAlerts(payload AlertmanagerPayload) {
	webhooks, _ := config.LoadWebhooks()

	type AlertGroup struct {
		WebhookURL  string
		Items       []notify.AlertItem
		FiringDescs []string
	}
	groups := make(map[string]*AlertGroup)

	for _, alert := range payload.Alerts {
		namespace := alert.Labels["namespace"]
		targetURL := s.DefaultWebhookURL
		if namespace != "" && webhooks != nil {
			if url, ok := webhooks[namespace]; ok && url != "" {
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

func (s *AgentServer) runAIAnalysis(targetWebhookURL string, descriptions []string) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("GEMINI_API_KEY is not set. Skipping AI analysis.")
		return
	}

	s.aiMutex.Lock()
	defer s.aiMutex.Unlock()

	time.Sleep(3 * time.Second)

	fmt.Println("Running AI analysis for fired alerts...")
	llmClient := gemini.NewClient(apiKey)

	alertContext := strings.Join(descriptions, "\n- ")
	prompt := fmt.Sprintf(`You are an expert DevOps AI assistant. 
The following Kubernetes alerts have just fired:
- %s

Provide a very short, concise diagnosis of the likely root cause and 1-2 bullet points for immediate troubleshooting steps. 
Keep the response under 150 words. Do not use markdown codeblocks that wrap the entire response.`, alertContext)

	diagnosis, err := llmClient.Analyze(prompt)
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
