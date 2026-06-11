package monitoring

// MetricsFetcher defines the interface for metrics monitoring integrations.
// Implementations of this interface retrieve resource usage statistics and status metrics
// from monitoring servers (e.g. Prometheus).
type MetricsFetcher interface {
	// Fetch retrieves the recent metrics and resource utilization stats for all workloads running in the specified Kubernetes namespace.
	Fetch(namespace string) (string, error)
}
