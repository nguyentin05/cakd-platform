package cluster

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/nguyentin05/cakd-platform/internal/config"
)

//go:embed k8s/*.yaml
var agentManifests embed.FS

// agentVars holds template variables used to dynamically populate the Kubernetes manifests of the CAKD Agent.
type agentVars struct {
	DiscordToken   string
	DiscordWebhook string
	GeminiKey      string
	Image          string
}

// deployAgent resolves local/environment credentials, renders the embedded Kubernetes
// yaml manifests for the CAKD Agent with the correct configuration, and applies them to the cluster.
func deployAgent(agentVersion string) error {
	discordToken, _, _ := config.GetDiscordCredentials()
	discordWebhook := os.Getenv("DISCORD_WEBHOOK_URL")
	geminiKey := config.GetGeminiAPIKey()

	if discordToken == "" || geminiKey == "" {
		fmt.Println("Warning: DISCORD_BOT_TOKEN or GEMINI_API_KEY is not set. Agent might not work fully.")
	}

	tag := "latest"
	if agentVersion != "" && agentVersion != "dev" {
		tag = agentVersion
	}

	vars := agentVars{
		DiscordToken:   discordToken,
		DiscordWebhook: discordWebhook,
		GeminiKey:      geminiKey,
		Image:          "ghcr.io/nguyentin05/cakd-agent:" + tag,
	}

	entries, err := agentManifests.ReadDir("k8s")
	if err != nil {
		return fmt.Errorf("failed to read k8s directory: %w", err)
	}

	var buf bytes.Buffer
	for _, entry := range entries {
		if entry.Name() == "prometheus-values.yaml" {
			continue
		}
		tmplData, err := agentManifests.ReadFile("k8s/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read embedded agent manifest %s: %w", entry.Name(), err)
		}

		tmpl, err := template.New(entry.Name()).Parse(string(tmplData))
		if err != nil {
			return fmt.Errorf("failed to parse agent template %s: %w", entry.Name(), err)
		}

		if err := tmpl.Execute(&buf, vars); err != nil {
			return fmt.Errorf("failed to execute agent template %s: %w", entry.Name(), err)
		}
		buf.WriteString("\n---\n")
	}

	fmt.Println("Step: Deploying CAKD Agent to Kubernetes...")
	kubectlCmd := exec.Command("kubectl", "apply", "-f", "-")
	kubectlCmd.Stdin = &buf
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr

	if err := kubectlCmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy agent manifests: %w", err)
	}

	return nil
}
