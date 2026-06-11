package iac

// Engine provisions infrastructure resources.
type Engine interface {
	Apply() (outputs map[string]string, err error)
	Destroy() error
}
