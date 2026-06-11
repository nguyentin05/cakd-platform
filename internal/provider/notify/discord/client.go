package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/provider/notify"
)

const discordAPIURL = "https://discord.com/api/v10"

type Client struct {
	token   string
	guildID string
	client  *http.Client
}

func NewClient(token, guildID string) *Client {
	return &Client{
		token:   token,
		guildID: guildID,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type ChannelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WebhookResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

// doRequest sends an authenticated HTTP request to the Discord API and decodes the response.
func (c *Client) doRequest(method, url string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) CreateChannel(projectName string) (string, error) {
	url := fmt.Sprintf("%s/guilds/%s/channels", discordAPIURL, c.guildID)
	payload := map[string]interface{}{
		"name": "alerts-" + projectName,
		"type": 0,
	}

	var resp ChannelResponse
	if err := c.doRequest("POST", url, payload, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) CreateWebhook(channelID string, projectName string) (string, error) {
	url := fmt.Sprintf("%s/channels/%s/webhooks", discordAPIURL, channelID)
	payload := map[string]interface{}{
		"name": "CAKD Alerting - " + projectName,
	}

	var resp WebhookResponse
	if err := c.doRequest("POST", url, payload, &resp); err != nil {
		return "", err
	}

	webhookURL := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", resp.ID, resp.Token)
	if resp.URL != "" {
		webhookURL = resp.URL
	}
	return webhookURL, nil
}

func (c *Client) ProvisionChannel(projectName string) (string, error) {
	fmt.Printf("   Creating Discord channel for %s...\n", projectName)
	channelID, err := c.CreateChannel(projectName)
	if err != nil {
		return "", fmt.Errorf("failed to create channel: %w", err)
	}

	fmt.Println("   Creating Discord webhook...")
	webhookURL, err := c.CreateWebhook(channelID, projectName)
	if err != nil {
		return "", fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhookURL, nil
}

type DiscordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type DiscordWebhookPayload struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

func (c *Client) SendAlert(webhookURL string, payload notify.AlertPayload) error {
	embeds := make([]DiscordEmbed, 0, len(payload.Items))

	for _, item := range payload.Items {
		color := 0x00FF00
		switch item.Severity {
		case "critical", "error", "firing":
			color = 0xFF0000
		case "warning":
			color = 0xFFA500
		}

		embeds = append(embeds, DiscordEmbed{
			Title:       item.Title,
			Description: item.Description,
			Color:       color,
			Timestamp:   time.Now().Format(time.RFC3339),
		})
	}

	if len(embeds) == 0 {
		return nil
	}

	msg := DiscordWebhookPayload{Embeds: embeds}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}
	return nil
}
