package agent

import "time"

// AlertmanagerPayload represents the incoming webhook payload format received from Prometheus Alertmanager.
type AlertmanagerPayload struct {
	Receiver string  `json:"receiver"`
	Status   string  `json:"status"`
	Alerts   []Alert `json:"alerts"`
	GroupKey string  `json:"groupKey"`
}

// Alert represents a single alert instance within the Alertmanager webhook payload.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}
