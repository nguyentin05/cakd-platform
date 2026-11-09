package catalogstore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

const (
	// DefaultRemoteURL is the remote repository endpoint for downloading official provider manifests.
	DefaultRemoteURL = "https://raw.githubusercontent.com/nguyentin05/cakd-platform/main/providers"
	// DefaultLocalDir is the repository-relative path checked first during development.
	DefaultLocalDir = "providers"
	// DefaultCacheTTL is the expiration threshold for local disk-cached provider manifests.
	DefaultCacheTTL = 24 * time.Hour
	// DefaultHTTPTimeout is the network timeout for catalog downloads and dynamic introspection.
	DefaultHTTPTimeout = 10 * time.Second
	// MaxResponseBytes is the maximum allowed response size (5MB) for catalog and introspection HTTP calls.
	MaxResponseBytes = 5 * 1024 * 1024
)

var (
	labelCandidates = []string{"name", "title", "label", "login", "slug", "key", "id", "value"}
	valueCandidates = []string{"value", "id", "key", "login", "slug", "name"}
)

// HTTPClient executes HTTP requests for remote provider manifests and introspection queries.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Store implements ports.ManifestLoader with local filesystem, ~/.cakd cache, and remote fetch resolution.
type Store struct {
	localDir   string
	cacheDir   string
	remoteURL  string
	cacheTTL   time.Duration
	httpClient HTTPClient
}

// Option provides functional options to configure Store.
type Option func(*Store)

// WithLocalDir overrides the local directory checked for manifests.
func WithLocalDir(dir string) Option {
	return func(s *Store) {
		s.localDir = dir
	}
}

// WithCacheDir overrides the cache directory.
func WithCacheDir(dir string) Option {
	return func(s *Store) {
		s.cacheDir = dir
	}
}

// WithRemoteURL overrides the remote base URL.
func WithRemoteURL(url string) Option {
	return func(s *Store) {
		s.remoteURL = strings.TrimRight(url, "/")
	}
}

// WithCacheTTL overrides cache TTL.
func WithCacheTTL(ttl time.Duration) Option {
	return func(s *Store) {
		s.cacheTTL = ttl
	}
}

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(client HTTPClient) Option {
	return func(s *Store) {
		s.httpClient = client
	}
}

// NewStore constructs a new catalog store driven adapter.
func NewStore(opts ...Option) *Store {
	s := &Store{
		localDir:   resolveLocalDir(DefaultLocalDir),
		cacheDir:   defaultCacheDir(),
		remoteURL:  DefaultRemoteURL,
		cacheTTL:   DefaultCacheTTL,
		httpClient: &http.Client{Timeout: DefaultHTTPTimeout},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

var _ ports.CatalogStore = (*Store)(nil)

// LoadIndex retrieves the catalog index following local -> cache -> remote fallback.
func (s *Store) LoadIndex(ctx context.Context) (*catalog.CatalogIndex, error) {
	// 1. Check local directory
	if s.localDir != "" {
		localPath := filepath.Join(s.localDir, "index.json")
		if data, err := os.ReadFile(localPath); err == nil {
			var idx catalog.CatalogIndex
			if err := yaml.Unmarshal(data, &idx); err != nil {
				return nil, errcode.Wrap(errcode.CodeValidation, "failed to parse local catalog index", err)
			}
			return &idx, nil
		}
	}

	// 2. Check cache
	cachePath := filepath.Join(s.cacheDir, "index.json")
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < s.cacheTTL {
			if data, err := os.ReadFile(cachePath); err == nil {
				var idx catalog.CatalogIndex
				if err := yaml.Unmarshal(data, &idx); err == nil {
					return &idx, nil
				}
			}
		}
	}

	// 3. Fetch remote
	remoteURL := s.remoteURL + "/index.json"
	data, fetchErr := s.fetchRemote(ctx, remoteURL)
	if fetchErr == nil {
		var idx catalog.CatalogIndex
		if err := yaml.Unmarshal(data, &idx); err != nil {
			return nil, errcode.Wrap(errcode.CodeValidation, "failed to parse remote catalog index", err)
		}
		s.saveToCache(cachePath, data)
		return &idx, nil
	}

	// 4. Stale cache fallback if remote fetch failed
	if data, err := os.ReadFile(cachePath); err == nil {
		var idx catalog.CatalogIndex
		if err := yaml.Unmarshal(data, &idx); err == nil {
			return &idx, nil
		}
	}

	if fetchErr != nil {
		return nil, errcode.Wrap(errcode.CodeProvision, "failed to load catalog index", fetchErr)
	}
	return nil, errcode.New(errcode.CodeNotFound, "failed to load catalog index: index not found")
}

// LoadManifest retrieves a provider manifest following local -> cache -> remote fallback.
func (s *Store) LoadManifest(ctx context.Context, category, providerID string) (*catalog.ProviderManifest, error) {
	category = strings.TrimSpace(category)
	providerID = strings.TrimSpace(providerID)
	if category == "" || providerID == "" || strings.Contains(category, "..") || strings.Contains(providerID, "..") {
		return nil, errcode.New(errcode.CodeValidation, "invalid category or provider ID: path traversal is forbidden")
	}

	if s.localDir != "" {
		if mf, found, err := s.readLocalManifest(category, providerID); found {
			return mf, err
		}
	}

	fileName := providerID + ".yaml"
	if mf, ok := s.readCacheManifest(category, fileName, false); ok {
		return mf, nil
	}

	remoteURL := fmt.Sprintf("%s/%s/%s", s.remoteURL, category, fileName)
	if mf, err := s.fetchAndCacheManifest(ctx, category, fileName, remoteURL); err == nil {
		return mf, nil
	} else if err != nil && errcode.IsCode(err, errcode.CodeValidation) {
		return nil, err
	}

	if mf, ok := s.readCacheManifest(category, fileName, true); ok {
		return mf, nil
	}

	return nil, errcode.New(errcode.CodeNotFound, fmt.Sprintf("manifest not found for %s/%s", category, providerID))
}

func (s *Store) readCacheManifest(category, fileName string, allowExpired bool) (*catalog.ProviderManifest, bool) {
	cachePath := filepath.Join(s.cacheDir, category, fileName)
	info, err := os.Stat(cachePath)
	if err != nil {
		return nil, false
	}
	if !allowExpired && time.Since(info.ModTime()) >= s.cacheTTL {
		return nil, false
	}
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}
	var m catalog.ProviderManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	return &m, true
}

func (s *Store) fetchAndCacheManifest(ctx context.Context, category, fileName, remoteURL string) (*catalog.ProviderManifest, error) {
	data, err := s.fetchRemote(ctx, remoteURL)
	if err != nil {
		return nil, err
	}
	var m catalog.ProviderManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, errcode.Wrap(errcode.CodeValidation, "failed to parse remote manifest", err)
	}
	s.saveToCache(filepath.Join(s.cacheDir, category, fileName), data)
	return &m, nil
}

func (s *Store) readLocalManifest(category, providerID string) (*catalog.ProviderManifest, bool, error) {
	candidates := []string{
		filepath.Join(s.localDir, category, providerID+".yaml"),
		filepath.Join(s.localDir, category, providerID+".yml"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			var m catalog.ProviderManifest
			if err := yaml.Unmarshal(data, &m); err != nil {
				return nil, true, errcode.Wrap(errcode.CodeValidation, "failed to parse manifest at "+path, err)
			}
			return &m, true, nil
		}
	}
	return nil, false, nil
}

// Introspect executes dynamic introspection queries against 3rd-party provider endpoints and evaluates JSONPath projections.
func (s *Store) Introspect(ctx context.Context, manifest *catalog.ProviderManifest, params map[string]string) ([]catalog.ChoiceOption, error) {
	if manifest == nil {
		return nil, errcode.New(errcode.CodeValidation, "manifest is nil")
	}

	targetStep, err := findTargetStep(manifest, params)
	if err != nil {
		return nil, err
	}

	if manifest.Introspection == nil || manifest.Introspection.Fetch.URL == "" {
		if targetStep != nil && len(targetStep.Options) > 0 {
			return targetStep.Options, nil
		}
		return nil, nil
	}

	req, fetchURL, err := s.buildIntrospectionRequest(ctx, manifest.Introspection, params)
	if err != nil {
		return nil, err
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: DefaultHTTPTimeout}
	}

	body, err := executeHTTPGet(client, req, fetchURL)
	if err != nil {
		return nil, err
	}

	var options []catalog.ChoiceOption
	if targetStep != nil && targetStep.Extract != "" {
		options = extractChoices(string(body), targetStep.Extract)
	}

	options = s.appendGitHubOrgs(ctx, client, req, fetchURL, options)

	if len(options) == 0 && targetStep != nil && len(targetStep.Options) > 0 {
		return targetStep.Options, nil
	}

	return options, nil
}

func (s *Store) buildIntrospectionRequest(ctx context.Context, intro *catalog.IntrospectionSpec, params map[string]string) (*http.Request, string, error) {
	fetchURL := interpolate(intro.Fetch.URL, params)
	method := intro.Fetch.Method
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, method, fetchURL, nil)
	if err != nil {
		return nil, "", errcode.Wrap(errcode.CodeProvision, "failed to create introspection request", err)
	}
	for k, v := range intro.Fetch.Headers {
		req.Header.Set(k, interpolate(v, params))
	}
	return req, fetchURL, nil
}

func executeHTTPGet(client HTTPClient, req *http.Request, fetchURL string) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, errcode.Wrap(errcode.CodeProvision, "introspection request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errcode.New(errcode.CodeNotFound, fmt.Sprintf("introspection endpoint not found (HTTP 404): %s", fetchURL))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errcode.New(errcode.CodeProvision, fmt.Sprintf("introspection endpoint returned HTTP %d", resp.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes))
	if err != nil {
		return nil, errcode.Wrap(errcode.CodeProvision, "failed to read introspection response", err)
	}
	return body, nil
}

func (s *Store) appendGitHubOrgs(ctx context.Context, client HTTPClient, origReq *http.Request, fetchURL string, options []catalog.ChoiceOption) []catalog.ChoiceOption {
	if !strings.Contains(fetchURL, "api.github.com/user") || strings.Contains(fetchURL, "/user/orgs") {
		return options
	}
	orgsURL := strings.Replace(fetchURL, "/user", "/user/orgs", 1)
	orgsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, orgsURL, nil)
	if err != nil {
		return options
	}
	for k, v := range origReq.Header {
		orgsReq.Header[k] = v
	}
	orgsResp, err := client.Do(orgsReq)
	if err != nil || orgsResp.StatusCode != http.StatusOK {
		return options
	}
	defer func() { _ = orgsResp.Body.Close() }()
	orgsBody, err := io.ReadAll(io.LimitReader(orgsResp.Body, MaxResponseBytes))
	if err != nil {
		return options
	}
	orgOptions := extractChoices(string(orgsBody), "$[*].login")
	return append(options, orgOptions...)
}

func (s *Store) fetchRemote(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errcode.Wrap(errcode.CodeProvision, "failed to create remote request", err)
	}
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errcode.Wrap(errcode.CodeProvision, "failed to fetch remote URL: "+url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errcode.New(errcode.CodeNotFound, fmt.Sprintf("remote resource not found (HTTP 404): %s", url))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errcode.New(errcode.CodeProvision, fmt.Sprintf("HTTP GET %s returned %d %s", url, resp.StatusCode, resp.Status))
	}

	return io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes))
}

func (s *Store) saveToCache(path string, data []byte) {
	if s.cacheDir == "" {
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o600)
}

func defaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "cakd-catalog-cache")
	}
	return filepath.Join(home, ".cakd", "catalog")
}

func resolveLocalDir(dir string) string {
	if dir == "" {
		return ""
	}
	if _, err := os.Stat(filepath.Join(dir, "index.json")); err == nil {
		return dir
	}
	cwd, err := os.Getwd()
	if err == nil {
		curr := cwd
		for {
			candidate := filepath.Join(curr, dir)
			if _, err := os.Stat(filepath.Join(candidate, "index.json")); err == nil {
				return candidate
			}
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}
	return dir
}

func findTargetStep(manifest *catalog.ProviderManifest, params map[string]string) (*catalog.StepSpec, error) {
	stepID := ""
	if params != nil {
		if s, ok := params["step"]; ok && s != "" {
			stepID = s
		} else if s, ok := params["step_id"]; ok && s != "" {
			stepID = s
		}
	}

	if stepID != "" {
		for i := range manifest.Steps {
			if manifest.Steps[i].ID == stepID {
				return &manifest.Steps[i], nil
			}
		}
		return nil, errcode.New(errcode.CodeNotFound, "requested step not found: "+stepID)
	}

	for i := range manifest.Steps {
		step := &manifest.Steps[i]
		if step.Extract != "" || len(step.Options) > 0 {
			return step, nil
		}
	}

	return nil, nil
}

func interpolate(text string, params map[string]string) string {
	if params == nil {
		return text
	}
	res := text
	for k, v := range params {
		res = strings.ReplaceAll(res, "{{."+k+"}}", v)
		res = strings.ReplaceAll(res, "{{"+k+"}}", v)
	}
	return res
}

func extractChoices(rawJSON, expr string) []catalog.ChoiceOption {
	expr = strings.TrimSpace(expr)
	var res gjson.Result
	if expr == "" || expr == "$" || expr == "$[*]" || expr == "@this" {
		res = gjson.Parse(rawJSON)
	} else {
		normalized := normalizePath(expr)
		res = gjson.Get(rawJSON, normalized)
		if !res.Exists() && normalized != expr {
			res = gjson.Get(rawJSON, expr)
		}
	}

	if !res.Exists() {
		return nil
	}

	if res.IsArray() {
		return extractFromArray(res.Array())
	}

	if res.IsObject() {
		opt := extractFromObject(res)
		if opt.Label != "" || opt.Value != "" {
			return []catalog.ChoiceOption{opt}
		}
		return nil
	}

	val := strings.TrimSpace(res.String())
	if val != "" {
		return []catalog.ChoiceOption{{Label: val, Value: val}}
	}
	return nil
}

func extractFromArray(items []gjson.Result) []catalog.ChoiceOption {
	var opts []catalog.ChoiceOption
	for _, item := range items {
		if item.IsObject() {
			opt := extractFromObject(item)
			if opt.Label != "" || opt.Value != "" {
				opts = append(opts, opt)
			}
		} else {
			val := strings.TrimSpace(item.String())
			if val != "" {
				opts = append(opts, catalog.ChoiceOption{Label: val, Value: val})
			}
		}
	}
	return opts
}

func extractFromObject(item gjson.Result) catalog.ChoiceOption {
	label := findFirstMatch(item, labelCandidates)
	val := findFirstMatch(item, valueCandidates)

	if label == "" {
		label = val
	}
	if val == "" {
		val = label
	}

	return catalog.ChoiceOption{
		Label: label,
		Value: val,
	}
}

func findFirstMatch(item gjson.Result, candidates []string) string {
	for _, key := range candidates {
		if v := item.Get(key); v.Exists() {
			s := strings.TrimSpace(v.String())
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func normalizePath(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "$[*].") {
		return "#." + strings.TrimPrefix(expr, "$[*].")
	}
	if expr == "$[*]" {
		return "#"
	}
	if strings.HasPrefix(expr, "$.") {
		return strings.TrimPrefix(expr, "$.")
	}
	return expr
}
