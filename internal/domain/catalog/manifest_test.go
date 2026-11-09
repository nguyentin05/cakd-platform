package catalog_test

import (
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
)

func TestCatalogManifestTypes(t *testing.T) {
	desc := catalog.ProviderDescriptor{
		ID:          "local-git",
		Name:        "Local Git",
		Category:    "vcs",
		Description: "Local repository without remote push",
		Version:     "1.0.0",
	}

	index := catalog.CatalogIndex{
		APIVersion: "v1",
		Providers:  []catalog.ProviderDescriptor{desc},
	}

	if index.APIVersion != "v1" {
		t.Fatalf("expected apiVersion v1, got %q", index.APIVersion)
	}
	if len(index.Providers) != 1 || index.Providers[0].ID != "local-git" {
		t.Fatalf("unexpected providers list: %v", index.Providers)
	}

	const optPrivate = "private"
	manifest := catalog.ProviderManifest{
		ID:          "github",
		Name:        "GitHub",
		Category:    "vcs",
		Description: "GitHub provider",
		Version:     "1.0.0",
		Introspection: &catalog.IntrospectionSpec{
			Fetch: catalog.FetchSpec{
				URL: "https://api.github.com/user/orgs",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
		},
		Steps: []catalog.StepSpec{
			{
				ID:      "owner",
				Title:   "Owner",
				Extract: "$[*].login",
			},
			{
				ID:      "visibility",
				Title:   "Visibility",
				Default: optPrivate,
				Options: []catalog.ChoiceOption{
					{Label: "Private", Value: optPrivate},
					{Label: "Public", Value: "public"},
				},
			},
		},
	}

	if manifest.Introspection == nil || manifest.Introspection.Fetch.URL != "https://api.github.com/user/orgs" {
		t.Fatalf("unexpected introspection URL: %v", manifest.Introspection)
	}
	if len(manifest.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(manifest.Steps))
	}
	if manifest.Steps[1].Options[0].Value != "private" {
		t.Fatalf("expected first option to be private, got %q", manifest.Steps[1].Options[0].Value)
	}
}
