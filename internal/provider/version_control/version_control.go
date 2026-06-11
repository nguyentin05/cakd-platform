package version_control

// VersionControl defines the interface for managing remote source code repositories.
// Implementations of this interface interact with remote VCS providers (e.g. GitHub).
type VersionControl interface {
	// InitAndPush initializes a local git repository in dir, commits all generated files,
	// and pushes the codebase to the specified remote repoURL using the provided auth token.
	InitAndPush(dir string, repoURL string, token string) error
}
