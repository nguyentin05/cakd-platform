package catalog

import (
	"context"
	"fmt"
	"strings"

	domaincatalog "github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// Service orchestrates provider catalog discovery, manifest resolution, and dynamic option introspection.
type Service struct {
	store ports.CatalogStore
}

// NewService constructs a new catalog UseCase Service.
func NewService(store ports.CatalogStore) *Service {
	return &Service{
		store: store,
	}
}

// ListProviders returns all registered providers, optionally filtered by category.
func (s *Service) ListProviders(ctx context.Context, category string) ([]domaincatalog.ProviderDescriptor, error) {
	if s.store == nil {
		return nil, errcode.New(errcode.CodeValidation, "catalog store is required")
	}
	idx, err := s.store.LoadIndex(ctx)
	if err != nil {
		return nil, err
	}
	if idx == nil {
		return nil, nil
	}
	if category == "" {
		return idx.Providers, nil
	}
	var filtered []domaincatalog.ProviderDescriptor
	for _, p := range idx.Providers {
		if strings.EqualFold(p.Category, category) {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// GetManifest loads the complete manifest for a specific provider.
func (s *Service) GetManifest(ctx context.Context, category, providerID string) (*domaincatalog.ProviderManifest, error) {
	if s.store == nil {
		return nil, errcode.New(errcode.CodeValidation, "catalog store is required")
	}
	return s.store.LoadManifest(ctx, category, providerID)
}

// Introspect queries dynamic configuration choices for a provider step via the catalog store.
func (s *Service) Introspect(ctx context.Context, category, providerID string, params map[string]string) ([]domaincatalog.ChoiceOption, error) {
	if s.store == nil {
		return nil, errcode.New(errcode.CodeValidation, "catalog store is required")
	}
	manifest, err := s.GetManifest(ctx, category, providerID)
	if err != nil {
		return nil, err
	}
	if manifest == nil {
		return nil, errcode.New(errcode.CodeProvision, fmt.Sprintf("manifest not found for %s/%s", category, providerID))
	}
	return s.store.Introspect(ctx, manifest, params)
}
