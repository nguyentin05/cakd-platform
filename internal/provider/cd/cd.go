package cd

// CD manages continuous deployment registration.
type CD interface {
	Register(manifestPath string) error
}
