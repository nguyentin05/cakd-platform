package monitoring

// MetricsFetcher retrieves metrics from a monitoring system.
type MetricsFetcher interface {
	Fetch(namespace string) (string, error)
}
