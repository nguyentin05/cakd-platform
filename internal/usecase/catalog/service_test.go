package catalog_test

import (
	"context"
	"testing"

	domaincatalog "github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	usecasecatalog "github.com/nguyentin05/cakd-platform/internal/usecase/catalog"
)

type mockStore struct {
	index          *domaincatalog.CatalogIndex
	manifests      map[string]*domaincatalog.ProviderManifest
	introspectFunc func(ctx context.Context, manifest *domaincatalog.ProviderManifest, params map[string]string) ([]domaincatalog.ChoiceOption, error)
}

func (m *mockStore) LoadIndex(_ context.Context) (*domaincatalog.CatalogIndex, error) {
	return m.index, nil
}

func (m *mockStore) LoadManifest(_ context.Context, category, providerID string) (*domaincatalog.ProviderManifest, error) {
	key := category + "/" + providerID
	if mf, ok := m.manifests[key]; ok {
		return mf, nil
	}
	return nil, nil
}

func (m *mockStore) Introspect(ctx context.Context, manifest *domaincatalog.ProviderManifest, params map[string]string) ([]domaincatalog.ChoiceOption, error) {
	if m.introspectFunc != nil {
		return m.introspectFunc(ctx, manifest, params)
	}
	return []domaincatalog.ChoiceOption{{Label: "opt1", Value: "val1"}}, nil
}

func TestListProviders(t *testing.T) {
	const catVCS = "vcs"
	store := &mockStore{
		index: &domaincatalog.CatalogIndex{
			APIVersion: "v1",
			Providers: []domaincatalog.ProviderDescriptor{
				{ID: "github", Category: catVCS},
				{ID: "local-git", Category: catVCS},
				{ID: "argocd", Category: "cd"},
			},
		},
	}

	svc := usecasecatalog.NewService(store)

	// List all
	all, err := svc.ListProviders(context.Background(), "")
	if err != nil {
		t.Fatalf("ListProviders all failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(all))
	}

	// Filter by category
	vcsList, err := svc.ListProviders(context.Background(), catVCS)
	if err != nil {
		t.Fatalf("ListProviders vcs failed: %v", err)
	}
	if len(vcsList) != 2 {
		t.Fatalf("expected 2 vcs providers, got %d", len(vcsList))
	}

	cdList, err := svc.ListProviders(context.Background(), "cd")
	if err != nil {
		t.Fatalf("ListProviders cd failed: %v", err)
	}
	if len(cdList) != 1 || cdList[0].ID != "argocd" {
		t.Fatalf("expected argocd in cd category, got %v", cdList)
	}
}

func TestGetManifest(t *testing.T) {
	store := &mockStore{
		manifests: map[string]*domaincatalog.ProviderManifest{
			"vcs/local-git": {
				ID:       "local-git",
				Category: "vcs",
				Name:     "Local Git",
			},
		},
	}

	svc := usecasecatalog.NewService(store)
	mf, err := svc.GetManifest(context.Background(), "vcs", "local-git")
	if err != nil {
		t.Fatalf("GetManifest failed: %v", err)
	}
	if mf == nil || mf.Name != "Local Git" {
		t.Fatalf("unexpected manifest: %+v", mf)
	}
}

func TestIntrospectDelegation(t *testing.T) {
	const (
		provGitHub  = "github"
		valAcmeCorp = "acme-corp"
	)
	called := false
	store := &mockStore{
		manifests: map[string]*domaincatalog.ProviderManifest{
			"vcs/github": {
				ID:       provGitHub,
				Category: "vcs",
			},
		},
		introspectFunc: func(_ context.Context, manifest *domaincatalog.ProviderManifest, params map[string]string) ([]domaincatalog.ChoiceOption, error) {
			called = true
			if manifest.ID != provGitHub {
				t.Fatalf("expected github manifest, got %s", manifest.ID)
			}
			if params["token"] != "test-token" {
				t.Fatalf("expected token in params, got %v", params)
			}
			return []domaincatalog.ChoiceOption{
				{Label: valAcmeCorp, Value: valAcmeCorp},
			}, nil
		},
	}

	svc := usecasecatalog.NewService(store)
	opts, err := svc.Introspect(context.Background(), "vcs", provGitHub, map[string]string{
		"token": "test-token",
	})
	if err != nil {
		t.Fatalf("Introspect failed: %v", err)
	}
	if !called {
		t.Fatal("expected store.Introspect to be called")
	}
	if len(opts) != 1 || opts[0].Value != valAcmeCorp {
		t.Fatalf("unexpected choices: %+v", opts)
	}
}

func TestIntrospectManifestNotFound(t *testing.T) {
	store := &mockStore{}
	svc := usecasecatalog.NewService(store)
	_, err := svc.Introspect(context.Background(), "vcs", "unknown", nil)
	if err == nil {
		t.Fatal("expected error when manifest not found")
	}
}
