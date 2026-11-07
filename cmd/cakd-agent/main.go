package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/adapter/webhook"
	"github.com/nguyentin05/cakd-platform/internal/composition"
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

	geminiKey := os.Getenv("GEMINI_API_KEY")
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-2.5-flash"
	}
	agentSecret := os.Getenv("CAKD_AGENT_SECRET")
	discordToken := os.Getenv("DISCORD_BOT_TOKEN")
	discordGuild := os.Getenv("DISCORD_GUILD_ID")

	dispatcher := composition.NewAlertDispatcher(webhookURL, discordToken, discordGuild, geminiKey, geminiModel)
	handler := webhook.NewHandler(dispatcher, agentSecret)

	addr := ":" + port
	fmt.Printf("CAKD Agent running on %s\n", addr)

	srv := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           handler.Routes(),
	}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
