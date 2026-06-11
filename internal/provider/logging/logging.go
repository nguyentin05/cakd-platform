package logging

// LogFetcher retrieves logs from a logging system.
type LogFetcher interface {
	Fetch(namespace string) (string, error)
}
