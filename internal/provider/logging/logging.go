package logging

// LogFetcher defines the interface for logs aggregation integrations.
// Implementations of this interface retrieve application logs from log aggregation backends (e.g. Loki).
type LogFetcher interface {
	// Fetch retrieves the recent system logs for all containers running in the specified Kubernetes namespace.
	Fetch(namespace string) (string, error)
}
