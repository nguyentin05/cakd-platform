package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client implements the [llm.LLM] interface for interacting with the Google Gemini API.
type Client struct {
	apiKey string
}

// NewClient initializes and returns a new Gemini API Client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// GeminiRequest represents the request body payload structure for the Gemini API.
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content holds a list of parts inside the Gemini Request structure.
type Content struct {
	Parts []Part `json:"parts"`
}

// Part holds the textual chunk payload.
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents the structured JSON response payload returned by the Gemini API.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []Part `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Analyze queries the Gemini API with the provided prompt text and returns the generated content diagnosis response.
func (c *Client) Analyze(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent?key=%s", c.apiKey)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	/* #nosec G107 */
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api return status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content returned from gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
