package loki

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// Client implements the [logging.LogFetcher] interface for Grafana Loki.
// It queries log streams from the Loki service running inside the Kubernetes cluster.
type Client struct{}

// NewClient initializes and returns a new Loki Client.
func NewClient() *Client {
	return &Client{}
}

// LokiResponse represents the structured query response format returned by the Grafana Loki API.
type LokiResponse struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// Fetch queries the Loki service inside the cluster (via kubectl proxy) to retrieve the recent
// 50 logs for all containers running inside the specified namespace.
func (c *Client) Fetch(namespace string) (string, error) {
	query := fmt.Sprintf(`{namespace="%s"}`, namespace)
	encodedQuery := url.QueryEscape(query)

	path := fmt.Sprintf("/api/v1/namespaces/monitoring/services/loki:3100/proxy/loki/api/v1/query_range?query=%s&limit=50", encodedQuery)

	cmd := exec.Command("kubectl", "get", "--raw", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query loki: %s", string(output))
	}

	var lokiResp LokiResponse
	if err := json.Unmarshal(output, &lokiResp); err != nil {
		return "", fmt.Errorf("failed to parse loki response: %w", err)
	}

	var result string
	result += fmt.Sprintf("Recent Logs for namespace %s:\n", namespace)

	for _, res := range lokiResp.Data.Result {
		pod := res.Stream["pod"]
		container := res.Stream["container"]
		result += fmt.Sprintf("--- Pod: %s | Container: %s ---\n", pod, container)

		for i := len(res.Values) - 1; i >= 0; i-- {
			logEntry := res.Values[i]
			if len(logEntry) == 2 {
				logStr := logEntry[1]
				logStr = strings.TrimSpace(logStr)
				result += logStr + "\n"
			}
		}
	}

	return result, nil
}
