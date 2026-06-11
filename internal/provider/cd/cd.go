package cd

// CD defines the interface for continuous deployment integrations.
// Implementations of this interface manage registering GitOps applications
// with external GitOps controllers (e.g. ArgoCD).
type CD interface {
	// Register registers a GitOps deployment application using the manifest file at the given path.
	Register(manifestPath string) error
}
