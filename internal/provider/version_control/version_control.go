package version_control

// VersionControl manages source code repositories.
type VersionControl interface {
	InitAndPush(dir string, repoURL string, token string) error
}
