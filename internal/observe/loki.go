package observe

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type LokiClient struct{}

func NewLokiClient() *LokiClient {
	return &LokiClient{}
}

type LokiResponse struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func (c *LokiClient) Fetch(namespace string) (string, error) {
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
