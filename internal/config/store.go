package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	storeMutex sync.Mutex
)

// getWebhooksFilePath resolves the path to the webhooks.json file.
// It prioritizes `/etc/cakd/webhooks.json` if it exists, and falls back
// to the user's home directory under `~/.cakd/webhooks.json`.
func getWebhooksFilePath() (string, error) {
	if _, err := os.Stat("/etc/cakd/webhooks.json"); err == nil {
		return "/etc/cakd/webhooks.json", nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	cakdDir := filepath.Join(homeDir, ".cakd")
	if err := os.MkdirAll(cakdDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .cakd directory: %w", err)
	}

	return filepath.Join(cakdDir, "webhooks.json"), nil
}

// LoadWebhooks loads and returns the map of project webhooks from the webhooks.json file.
// If the file does not exist, it returns an empty map without error.
func LoadWebhooks() (map[string]string, error) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	filePath, err := getWebhooksFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhooks file: %w", err)
	}

	webhooks := make(map[string]string)
	if err := json.Unmarshal(data, &webhooks); err != nil {
		return nil, fmt.Errorf("failed to parse webhooks json: %w", err)
	}

	return webhooks, nil
}

// SaveWebhook records a Discord webhook URL for a project/namespace.
// It saves the webhook to the local store and attempts to synchronize it
// with the Kubernetes cluster by patching the 'cakd-webhooks' ConfigMap.
func SaveWebhook(projectName, webhookURL string) error {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	filePath, err := getWebhooksFilePath()
	if err != nil {
		return err
	}

	webhooks := make(map[string]string)

	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err == nil {
			_ = json.Unmarshal(data, &webhooks)
		}
	}

	webhooks[projectName] = webhookURL

	data, err := json.MarshalIndent(webhooks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal webhooks: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write webhooks file: %w", err)
	}

	patchData := fmt.Sprintf(`{"data":{"webhooks.json":%q}}`, string(data))
	cmd := exec.Command("kubectl", "patch", "configmap", "cakd-webhooks", "-n", "monitoring", "-p", patchData)
	if err := cmd.Run(); err != nil {
		fmt.Printf("   Warning: Could not sync webhook to Kubernetes ConfigMap (is the cluster running?)\n")
	}

	return nil
}
