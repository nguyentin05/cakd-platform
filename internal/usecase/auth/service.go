package auth

import "github.com/nguyentin05/cakd-platform/internal/usecase/ports"

// Service manages local provider credentials via ports.CredentialStore.
type Service struct {
	store ports.CredentialStore
}

// NewService creates an auth Service.
func NewService(store ports.CredentialStore) *Service {
	return &Service{store: store}
}

// Set persists non-empty credentials in creds.
func (s *Service) Set(creds ports.CredentialSet) error {
	return s.store.Save(creds)
}

// Get retrieves resolved credentials (environment variables taking precedence over stored file).
func (s *Service) Get() (ports.CredentialSet, error) {
	return s.store.Get()
}

// Status returns a masked summary of configured credentials.
func (s *Service) Status() (ports.CredentialStatus, error) {
	return s.store.Status()
}

// Logout removes stored credentials for provider ("github", "discord", "gemini", or "all").
func (s *Service) Logout(provider string) error {
	return s.store.Delete(provider)
}

// Path returns the credentials file path on disk.
func (s *Service) Path() string {
	return s.store.Path()
}
