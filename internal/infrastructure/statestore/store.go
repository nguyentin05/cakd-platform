package statestore

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// Store implements ports.ClusterStateStore backed by ~/.cakd (or /etc/cakd) JSON files.
type Store struct {
	mu      sync.Mutex
	baseDir string
}

// NewStore creates a Store using /etc/cakd if present, or ~/.cakd.
func NewStore() *Store {
	if st, err := os.Stat("/etc/cakd/webhooks.json"); err == nil && !st.IsDir() {
		return &Store{baseDir: "/etc/cakd"}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return &Store{baseDir: ".cakd"}
	}
	return &Store{baseDir: filepath.Join(home, ".cakd")}
}

// NewStoreWithDir creates a Store rooted at a specific directory (useful for tests).
func NewStoreWithDir(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

func (s *Store) clusterStatePath() string {
	return filepath.Join(s.baseDir, "cluster-state.json")
}

func (s *Store) webhooksPath() string {
	return filepath.Join(s.baseDir, "webhooks.json")
}

// SaveClusterState persists the active cluster state (host, domain, kubeconfigPath) to disk.
func (s *Store) SaveClusterState(state ports.ClusterState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.baseDir, 0o700); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to create state directory", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to serialize cluster state", err)
	}
	if err := os.WriteFile(s.clusterStatePath(), data, 0o600); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to write cluster state file", err)
	}
	return nil
}

// LoadClusterState reads the persisted cluster state, returning an empty ClusterState if none exists.
func (s *Store) LoadClusterState() (ports.ClusterState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var state ports.ClusterState
	data, err := os.ReadFile(s.clusterStatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, errcode.Wrap(errcode.CodeExternal, "failed to read cluster state file", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, errcode.Wrap(errcode.CodeValidation, "invalid cluster state JSON", err)
	}
	return state, nil
}

// LoadWebhooks reads the map of project/namespace names to Discord webhook URLs.
func (s *Store) LoadWebhooks() (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.loadWebhooksUnlocked()
}

func (s *Store) loadWebhooksUnlocked() (map[string]string, error) {
	hooks := make(map[string]string)
	data, err := os.ReadFile(s.webhooksPath())
	if err != nil {
		if os.IsNotExist(err) {
			return hooks, nil
		}
		return nil, errcode.Wrap(errcode.CodeExternal, "failed to read webhooks file", err)
	}
	if err := json.Unmarshal(data, &hooks); err != nil {
		return nil, errcode.Wrap(errcode.CodeValidation, "invalid webhooks JSON", err)
	}
	return hooks, nil
}

// SaveWebhook records a webhook URL for projectName locally and best-effort syncs it to the cluster ConfigMap.
func (s *Store) SaveWebhook(ctx context.Context, projectName, webhookURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.baseDir, 0o700); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to create state directory", err)
	}
	hooks, err := s.loadWebhooksUnlocked()
	if err != nil {
		hooks = make(map[string]string)
	}
	hooks[projectName] = webhookURL

	data, err := json.MarshalIndent(hooks, "", "  ")
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to marshal webhooks", err)
	}
	if err := os.WriteFile(s.webhooksPath(), data, 0o600); err != nil {
		return errcode.Wrap(errcode.CodeExternal, "failed to write webhooks file", err)
	}
	return nil
}
