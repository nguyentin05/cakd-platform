package agent

import "time"

type AlertmanagerPayload struct {
	Receiver string  `json:"receiver"`
	Status   string  `json:"status"`
	Alerts   []Alert `json:"alerts"`
	GroupKey string  `json:"groupKey"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}
