package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/agent"
	"github.com/nguyentin05/cakd-platform/internal/provider/notify/discord"
)

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

	notifier := discord.NewClient("", "")
	server := agent.NewAgentServer(webhookURL, notifier)

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
