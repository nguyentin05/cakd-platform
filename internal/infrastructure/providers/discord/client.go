package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

const discordAPIURL = "https://discord.com/api/v10"

// Client implements ports.NotificationProvider for Discord channels and webhooks.
type Client struct {
	token      string
	guildID    string
	credStore  ports.CredentialStore
	apiBaseURL string
	httpClient *http.Client
}

// NewClient creates a Discord notification Client with static credentials.
func NewClient(token, guildID string) *Client {
	return &Client{
		token:      token,
		guildID:    guildID,
		apiBaseURL: discordAPIURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewClientWithStore creates a Discord notification Client that dynamically resolves credentials.
func NewClientWithStore(store ports.CredentialStore) *Client {
	return &Client{
		credStore:  store,
		apiBaseURL: discordAPIURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewClientWithStoreAndURL creates a Discord Client with a custom base URL (useful for testing).
func NewClientWithStoreAndURL(store ports.CredentialStore, apiBaseURL string) *Client {
	return &Client{
		credStore:  store,
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) resolveCredentials() (string, string) {
	if c.credStore != nil {
		if creds, err := c.credStore.Get(); err == nil {
			token := creds.DiscordBotToken
			guildID := creds.DiscordGuildID
			if token == "" {
				token = c.token
			}
			if guildID == "" {
				guildID = c.guildID
			}
			return token, guildID
		}
	}
	return c.token, c.guildID
}

type channelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type webhookResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

type embedItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type webhookPayload struct {
	Content string      `json:"content,omitempty"`
	Embeds  []embedItem `json:"embeds,omitempty"`
}

// ProvisionWebhook creates a Discord channel and webhook for channelName and returns the webhook URL.
func (c *Client) ProvisionWebhook(ctx context.Context, channelName string) (string, error) {
	token, guildID := c.resolveCredentials()
	if token == "" || guildID == "" {
		return "", errcode.New(errcode.CodeAuth, "Discord bot token and guild ID are required to provision webhooks")
	}

	var chResp channelResponse
	chURL := fmt.Sprintf("%s/guilds/%s/channels", c.apiBaseURL, guildID)
	err := c.doBotRequest(ctx, http.MethodPost, chURL, token, map[string]any{
		"name": "alerts-" + channelName,
		"type": 0,
	}, &chResp)
	if err != nil {
		return "", err
	}

	var whResp webhookResponse
	whURL := fmt.Sprintf("%s/channels/%s/webhooks", c.apiBaseURL, chResp.ID)
	err = c.doBotRequest(ctx, http.MethodPost, whURL, token, map[string]any{
		"name": "CAKD Alerting - " + channelName,
	}, &whResp)
	if err != nil {
		return "", err
	}

	if whResp.URL != "" {
		return whResp.URL, nil
	}
	return fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", whResp.ID, whResp.Token), nil
}

// SendAlert posts structured notification items as rich embeds to webhookURL.
func (c *Client) SendAlert(ctx context.Context, webhookURL string, items []diagnostic.NotificationItem) error {
	if len(items) == 0 || webhookURL == "" {
		return nil
	}
	embeds := make([]embedItem, 0, len(items))
	for _, item := range items {
		color := 0x00FF00
		switch item.Severity {
		case "critical", "error", "firing":
			color = 0xFF0000
		case "warning":
			color = 0xFFA500
		}
		embeds = append(embeds, embedItem{
			Title:       item.Title,
			Description: item.Description,
			Color:       color,
			Timestamp:   time.Now().Format(time.RFC3339),
		})
	}
	return c.postWebhook(ctx, webhookURL, webhookPayload{Embeds: embeds})
}

// SendMessage posts plain or markdown text to webhookURL.
func (c *Client) SendMessage(ctx context.Context, webhookURL, content string) error {
	if webhookURL == "" || content == "" {
		return nil
	}
	return c.postWebhook(ctx, webhookURL, webhookPayload{Content: content})
}

func (c *Client) postWebhook(ctx context.Context, webhookURL string, payload webhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "marshal discord webhook payload", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "create discord webhook request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "send discord webhook", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errcode.New(errcode.CodeExternal, fmt.Sprintf("discord webhook returned status %d", resp.StatusCode))
	}
	return nil
}

func (c *Client) doBotRequest(ctx context.Context, method, urlStr, token string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errcode.Wrap(errcode.CodeExternal, "discord API request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return errcode.New(errcode.CodeExternal, fmt.Sprintf("discord API error %d: %s", resp.StatusCode, string(respBody)))
	}
	return json.NewDecoder(resp.Body).Decode(result)
}
