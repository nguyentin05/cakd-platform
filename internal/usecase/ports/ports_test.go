package ports_test

import (
	"context"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

type mockManifestLoader struct {
	index    *catalog.CatalogIndex
	manifest *catalog.ProviderManifest
}

func (m *mockManifestLoader) LoadIndex(_ context.Context) (*catalog.CatalogIndex, error) {
	return m.index, nil
}

func (m *mockManifestLoader) LoadManifest(_ context.Context, _, _ string) (*catalog.ProviderManifest, error) {
	return m.manifest, nil
}

func (m *mockManifestLoader) Introspect(_ context.Context, _ *catalog.ProviderManifest, _ map[string]string) ([]catalog.ChoiceOption, error) {
	return []catalog.ChoiceOption{{Label: "opt1", Value: "val1"}}, nil
}

func TestManifestLoaderPort(t *testing.T) {
	var loader ports.ManifestLoader = &mockManifestLoader{
		index: &catalog.CatalogIndex{
			APIVersion: "v1",
			Providers: []catalog.ProviderDescriptor{
				{ID: catalog.ProviderGitHub, Category: "vcs"},
			},
		},
		manifest: &catalog.ProviderManifest{
			ID:       catalog.ProviderGitHub,
			Category: "vcs",
		},
	}

	idx, err := loader.LoadIndex(context.Background())
	if err != nil || idx == nil || len(idx.Providers) != 1 {
		t.Fatalf("unexpected index result: %v, err: %v", idx, err)
	}

	mf, err := loader.LoadManifest(context.Background(), "vcs", catalog.ProviderGitHub)
	if err != nil || mf == nil || mf.ID != catalog.ProviderGitHub {
		t.Fatalf("unexpected manifest result: %v, err: %v", mf, err)
	}
}
