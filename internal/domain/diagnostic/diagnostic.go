package diagnostic

import "time"

// AlertPayload represents the webhook payload sent by Prometheus Alertmanager.
type AlertPayload struct {
	Receiver string  `json:"receiver"`
	Status   string  `json:"status"`
	Alerts   []Alert `json:"alerts"`
	GroupKey string  `json:"groupKey"`
}

// Alert represents a single Alertmanager alert entry.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// NotificationItem represents a formatted alert item dispatched to a notification channel.
type NotificationItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// Report summarizes metrics, logs, and AI root-cause analysis for a workload.
type Report struct {
	ProjectName    string `json:"projectName"`
	Namespace      string `json:"namespace"`
	MetricsSummary string `json:"metricsSummary"`
	LogsSummary    string `json:"logsSummary"`
	Analysis       string `json:"analysis"`
	Notified       bool   `json:"notified"`
}
