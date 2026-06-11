package providers

const (
	Prometheus = "prometheus"
)

// VersionControl is the list of supported Version Control System providers.
var VersionControl = []string{"github"}

// CI is the list of supported Continuous Integration providers.
var CI = []string{"github-actions"}

// CD is the list of supported Continuous Deployment providers.
var CD = []string{"argocd"}

// Notification is the list of supported chat/alert notification channel providers.
var Notification = []string{"discord"}

// LLM is the list of supported Large Language Model providers for AI analysis.
var LLM = []string{"gemini"}

// Monitoring is the list of supported metrics collection and monitoring providers.
var Monitoring = []string{Prometheus}

// Logging is the list of supported log aggregation and logging providers.
var Logging = []string{"loki"}

// MonitoringDeps maps monitoring providers to their Spring Boot starter dependencies.
var MonitoringDeps = map[string]string{
	Prometheus: Prometheus,
}
