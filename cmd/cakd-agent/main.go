package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/agent"
	"github.com/nguyentin05/cakd-platform/internal/config"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify/discord"
)

// main is the entry point for the CAKD Agent server.
// It parses the Discord webhook configurations and starts an HTTP server listening for Alertmanager webhooks.
func main() {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Fprintln(os.Stderr, "Error: DISCORD_WEBHOOK_URL environment variable is required")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-flash-latest"
	}
	agentSecret := os.Getenv("CAKD_AGENT_SECRET")

	webhooks, err := config.LoadWebhooks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load webhooks: %v\n", err)
	}

	notifier := discord.NewClient("", "")
	server := agent.NewAgentServer(webhookURL, notifier, geminiKey, geminiModel, agentSecret, webhooks)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/alerts", server.HandleAlert)

	addr := ":" + port
	fmt.Printf("CAKD Agent running on %s\n", addr)
	fmt.Printf("Listening for Alertmanager webhooks at http://localhost%s/api/v1/alerts\n", addr)
	fmt.Println("Waiting for alerts...")

	srv := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
