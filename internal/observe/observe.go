package observe

import (
	"fmt"
)

type MetricsFetcher interface {
	Fetch(namespace string) (string, error)
}

type LogFetcher interface {
	Fetch(namespace string) (string, error)
}

type AIAnalyzer interface {
	Analyze(prompt string) (string, error)
}

type ObserverService struct {
	metrics MetricsFetcher
	logs    LogFetcher
	ai      AIAnalyzer
}

func NewObserverService(m MetricsFetcher, l LogFetcher, ai AIAnalyzer) *ObserverService {
	return &ObserverService{
		metrics: m,
		logs:    l,
		ai:      ai,
	}
}

func (s *ObserverService) Diagnose(namespace string) error {
	fmt.Println("Fetching metrics from Prometheus...")
	metricsData, err := s.metrics.Fetch(namespace)
	if err != nil {
		fmt.Printf("Warning: Could not fetch metrics: %v\n", err)
		metricsData = "No metrics available."
	}

	fmt.Println("Fetching logs from Loki...")
	logsData, err := s.logs.Fetch(namespace)
	if err != nil {
		fmt.Printf("Warning: Could not fetch logs: %v\n", err)
		logsData = "No logs available."
	}

	prompt := fmt.Sprintf(`You are an expert DevOps engineer and Kubernetes administrator.
The user is experiencing issues with their application in the "%s" namespace.
I have collected the following observability data from their cluster.

--- METRICS ---
%s

--- LOGS ---
%s

Please analyze this data and provide a diagnosis.
1. Identify if there are any pods crashing or restarting.
2. Look at the logs to determine the root cause of the failure.
3. Provide a clear, actionable solution to fix the problem.

Format your response in Markdown, using clear headings. Keep the explanation concise and focused on the root cause. Answer in Vietnamese.
`, namespace, metricsData, logsData)

	fmt.Println("Sending data to AI for diagnosis...")

	diagnosis, err := s.ai.Analyze(prompt)
	if err != nil {
		return fmt.Errorf("ai analysis failed: %w", err)
	}

	fmt.Println("\n==========================================")
	fmt.Println("CAKD AI DIAGNOSIS")
	fmt.Println("==========================================")
	fmt.Println(diagnosis)
	fmt.Println("==========================================")

	return nil
}
