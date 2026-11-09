package catalogstore_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/catalogstore"
)

func TestCatalogStoreLocalRead(t *testing.T) {
	const provGitHub = "github"
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, "providers")
	_ = os.MkdirAll(filepath.Join(localDir, "vcs"), 0o750)

	indexJSON := `{
  "apiVersion": "v1",
  "providers": [
    {"id": "github", "name": "GitHub", "category": "vcs"}
  ]
}`
	_ = os.WriteFile(filepath.Join(localDir, "index.json"), []byte(indexJSON), 0o600)

	manifestYAML := `id: github
name: GitHub
category: vcs
version: 1.0.0
steps:
  - id: owner
    title: Owner
`
	_ = os.WriteFile(filepath.Join(localDir, "vcs", "github.yaml"), []byte(manifestYAML), 0o600)

	store := catalogstore.NewStore(
		catalogstore.WithLocalDir(localDir),
		catalogstore.WithCacheDir(filepath.Join(tmp, "cache")),
	)

	idx, err := store.LoadIndex(context.Background())
	if err != nil {
		t.Fatalf("LoadIndex failed: %v", err)
	}
	if len(idx.Providers) != 1 || idx.Providers[0].ID != provGitHub {
		t.Fatalf("unexpected providers in index: %v", idx.Providers)
	}

	mf, err := store.LoadManifest(context.Background(), "vcs", provGitHub)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}
	if mf.ID != provGitHub || mf.Category != "vcs" || len(mf.Steps) != 1 {
		t.Fatalf("unexpected manifest content: %+v", mf)
	}
}

func TestCatalogStoreCacheAndRemoteFallback(t *testing.T) {
	const provGitLab = "gitlab"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
  "apiVersion": "v1",
  "providers": [
    {"id": "gitlab", "name": "GitLab", "category": "vcs"}
  ]
}`))
		case "/vcs/gitlab.yaml":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`id: gitlab
name: GitLab
category: vcs
`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	tmp := t.TempDir()
	cacheDir := filepath.Join(tmp, "cache")

	store := catalogstore.NewStore(
		catalogstore.WithLocalDir("nonexistent"),
		catalogstore.WithCacheDir(cacheDir),
		catalogstore.WithRemoteURL(ts.URL),
		catalogstore.WithHTTPClient(ts.Client()),
	)

	// Fetch from remote and populate cache
	idx, err := store.LoadIndex(context.Background())
	if err != nil {
		t.Fatalf("remote LoadIndex failed: %v", err)
	}
	if len(idx.Providers) != 1 || idx.Providers[0].ID != provGitLab {
		t.Fatalf("unexpected provider: %v", idx.Providers)
	}

	// Verify cache file was written
	if _, err := os.Stat(filepath.Join(cacheDir, "index.json")); err != nil {
		t.Fatalf("expected cache file for index.json: %v", err)
	}

	mf, err := store.LoadManifest(context.Background(), "vcs", provGitLab)
	if err != nil {
		t.Fatalf("remote LoadManifest failed: %v", err)
	}
	if mf.ID != provGitLab {
		t.Fatalf("expected gitlab manifest, got %s", mf.ID)
	}

	// Now shutdown remote server to verify cache is used
	ts.Close()

	idxCached, err := store.LoadIndex(context.Background())
	if err != nil {
		t.Fatalf("cached LoadIndex failed: %v", err)
	}
	if len(idxCached.Providers) != 1 || idxCached.Providers[0].ID != provGitLab {
		t.Fatalf("unexpected cached provider: %v", idxCached.Providers)
	}

	mfCached, err := store.LoadManifest(context.Background(), "vcs", provGitLab)
	if err != nil {
		t.Fatalf("cached LoadManifest failed: %v", err)
	}
	if mfCached.ID != provGitLab {
		t.Fatalf("expected gitlab manifest from cache, got %s", mfCached.ID)
	}
}

func TestCatalogStoreStaleCacheFallbackOnRemoteError(t *testing.T) {
	tmp := t.TempDir()
	cacheDir := filepath.Join(tmp, "cache")
	_ = os.MkdirAll(filepath.Join(cacheDir, "vcs"), 0o750)

	manifestPath := filepath.Join(cacheDir, "vcs", "cached-prov.yaml")
	_ = os.WriteFile(manifestPath, []byte("id: cached-prov\nname: Cached\ncategory: vcs\n"), 0o600)

	// Artificially age the file to make it expired
	oldTime := time.Now().Add(-48 * time.Hour)
	_ = os.Chtimes(manifestPath, oldTime, oldTime)

	// Store with short TTL and invalid remote URL
	store := catalogstore.NewStore(
		catalogstore.WithLocalDir("nonexistent"),
		catalogstore.WithCacheDir(cacheDir),
		catalogstore.WithCacheTTL(1*time.Hour),
		catalogstore.WithRemoteURL("http://127.0.0.1:1"), // unreachable
	)

	mf, err := store.LoadManifest(context.Background(), "vcs", "cached-prov")
	if err != nil {
		t.Fatalf("expected fallback to stale cache on remote error, got err: %v", err)
	}
	if mf.ID != "cached-prov" {
		t.Fatalf("expected cached-prov, got %s", mf.ID)
	}
}

func TestCatalogStorePathTraversalRejection(t *testing.T) {
	store := catalogstore.NewStore()
	_, err := store.LoadManifest(context.Background(), "../secret", "id")
	if err == nil {
		t.Fatalf("expected path traversal error for category")
	}

	_, err = store.LoadManifest(context.Background(), "vcs", "../../etc/passwd")
	if err == nil {
		t.Fatalf("expected path traversal error for providerID")
	}
}

func TestCatalogStoreUnmarshalErrorPropagation(t *testing.T) {
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, "providers")
	_ = os.MkdirAll(filepath.Join(localDir, "vcs"), 0o750)

	_ = os.WriteFile(filepath.Join(localDir, "vcs", "bad.yaml"), []byte("not valid yaml: [unclosed"), 0o600)

	store := catalogstore.NewStore(
		catalogstore.WithLocalDir(localDir),
	)

	_, err := store.LoadManifest(context.Background(), "vcs", "bad")
	if err == nil {
		t.Fatalf("expected unmarshal error to be propagated, got nil")
	}
}

func TestCatalogStoreIntrospect(t *testing.T) {
	const extractAll = "$[*]"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/orgs":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"slug": "core-team", "name": "Core Team"},
				{"slug": "infra-team", "name": "Infra Team"}
			]`))
		case "/primitives":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`["us-east-1", "eu-west-1", "ap-southeast-1"]`))
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	store := catalogstore.NewStore(
		catalogstore.WithHTTPClient(ts.Client()),
	)

	manifest := &catalog.ProviderManifest{
		ID:       "github",
		Category: "vcs",
		Introspection: &catalog.IntrospectionSpec{
			Fetch: catalog.FetchSpec{
				URL: ts.URL + "/orgs",
				Headers: map[string]string{
					"Authorization": "Bearer {{.token}}",
				},
			},
		},
		Steps: []catalog.StepSpec{
			{
				ID:      "owner",
				Title:   "Owner",
				Extract: "$[*].slug",
			},
			{
				ID:      "other",
				Title:   "Other",
				Extract: extractAll,
			},
		},
	}

	opts, err := store.Introspect(context.Background(), manifest, map[string]string{
		"token": "secret-123",
	})
	if err != nil {
		t.Fatalf("Introspect failed: %v", err)
	}
	if len(opts) != 2 || opts[0].Value != "core-team" || opts[1].Value != "infra-team" {
		t.Fatalf("unexpected options: %+v", opts)
	}

	// Test primitives array
	manifestPrimitives := &catalog.ProviderManifest{
		ID:       "cloud",
		Category: "infra",
		Introspection: &catalog.IntrospectionSpec{
			Fetch: catalog.FetchSpec{
				URL: ts.URL + "/primitives",
			},
		},
		Steps: []catalog.StepSpec{
			{ID: "regions", Extract: extractAll},
		},
	}

	primOpts, err := store.Introspect(context.Background(), manifestPrimitives, nil)
	if err != nil {
		t.Fatalf("Introspect primitives failed: %v", err)
	}
	if len(primOpts) != 3 || primOpts[0].Value != "us-east-1" {
		t.Fatalf("unexpected primitive options: %+v", primOpts)
	}

	// Test explicit step not found returns CodeNotFound
	_, err = store.Introspect(context.Background(), manifest, map[string]string{
		"step": "nonexistent_step",
	})
	if err == nil {
		t.Fatalf("expected error for nonexistent explicit step")
	}

	// Test 404 endpoint returns error
	manifest404 := &catalog.ProviderManifest{
		Introspection: &catalog.IntrospectionSpec{
			Fetch: catalog.FetchSpec{URL: ts.URL + "/notfound"},
		},
		Steps: []catalog.StepSpec{{ID: "s1", Extract: extractAll}},
	}
	_, err = store.Introspect(context.Background(), manifest404, nil)
	if err == nil {
		t.Fatalf("expected 404 error from introspection endpoint")
	}
}
