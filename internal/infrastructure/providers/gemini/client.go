package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
)

const defaultEndpoint = "https://generativelanguage.googleapis.com"

// Client implements ports.AIAnalyzer using Google Gemini's generateContent REST API.
type Client struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a Gemini AIAnalyzer Client.
func NewClient(apiKey, model string) *Client {
	return NewClientWithEndpoint(apiKey, model, defaultEndpoint)
}

// NewClientWithEndpoint creates a Gemini AIAnalyzer Client with a custom base endpoint (useful for tests).
func NewClientWithEndpoint(apiKey, model, endpoint string) *Client {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	return &Client{
		apiKey:     apiKey,
		model:      model,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

type generateRequest struct {
	Contents []contentBlock `json:"contents"`
}

type contentBlock struct {
	Parts []partBlock `json:"parts"`
}

type partBlock struct {
	Text string `json:"text"`
}

type generateResponse struct {
	Candidates []struct {
		Content contentBlock `json:"content"`
	} `json:"candidates"`
}

// Analyze sends prompt to the configured Gemini model and returns the generated diagnostic text.
func (c *Client) Analyze(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", errcode.New(errcode.CodeAuth, "Gemini API key is required for AI analysis")
	}
	reqURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.endpoint, c.model)
	payload := generateRequest{
		Contents: []contentBlock{
			{Parts: []partBlock{{Text: prompt}}},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "marshal gemini request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "create gemini request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "call gemini API", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "read gemini response", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", errcode.New(errcode.CodeExternal, fmt.Sprintf("gemini API returned status %d: %s", resp.StatusCode, string(respBytes)))
	}

	var parsed generateResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "decode gemini response", err)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", errcode.New(errcode.CodeExternal, "empty response from gemini API")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
