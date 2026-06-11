package providers

const (
	Prometheus = "prometheus"
)

var VersionControl = []string{"github"}
var CI = []string{"github-actions"}
var CD = []string{"argocd"}
var Notification = []string{"discord"}
var LLM = []string{"gemini"}
var Monitoring = []string{Prometheus}
var Logging = []string{"loki"}

// MonitoringDeps maps monitoring providers to their Spring Boot starter dependencies.
var MonitoringDeps = map[string]string{
	Prometheus: Prometheus,
}
