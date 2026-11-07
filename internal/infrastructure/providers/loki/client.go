package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
)

// Client implements ports.LogCollector for Grafana Loki (in-cluster HTTP or kubectl raw proxy).
type Client struct {
	httpClient *http.Client
}

// NewClient creates a Loki LogCollector Client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type queryResponse struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// CollectLogsSummary fetches recent container logs for namespace from Grafana Loki.
func (c *Client) CollectLogsSummary(ctx context.Context, namespace string) (string, error) {
	query := fmt.Sprintf(`{namespace="%s"}`, namespace)
	encoded := url.QueryEscape(query)

	var raw []byte
	var err error
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		apiURL := fmt.Sprintf("http://loki.monitoring.svc.cluster.local:3100/loki/api/v1/query_range?query=%s&limit=50", encoded)
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if reqErr != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "build loki request", reqErr)
		}
		resp, doErr := c.httpClient.Do(req)
		if doErr != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "query loki in-cluster", doErr)
		}
		defer resp.Body.Close()
		raw, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "read loki response", err)
		}
		if resp.StatusCode != http.StatusOK {
			return "", errcode.New(errcode.CodeExternal, fmt.Sprintf("loki status %d: %s", resp.StatusCode, string(raw)))
		}
	} else {
		path := fmt.Sprintf("/api/v1/namespaces/monitoring/services/loki:3100/proxy/loki/api/v1/query_range?query=%s&limit=50", encoded)
		//nolint:gosec
		cmd := exec.CommandContext(ctx, "kubectl", "get", "--raw", path)
		raw, err = cmd.CombinedOutput()
		if err != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "query loki via kubectl: "+strings.TrimSpace(string(raw)), err)
		}
	}

	var parsed queryResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "decode loki response", err)
	}

	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "Recent Logs for namespace %s:\n", namespace)
	for _, r := range parsed.Data.Result {
		_, _ = fmt.Fprintf(&sb, "--- Pod: %s | Container: %s ---\n", r.Stream["pod"], r.Stream["container"])
		for i := len(r.Values) - 1; i >= 0; i-- {
			if len(r.Values[i]) == 2 {
				sb.WriteString(strings.TrimSpace(r.Values[i][1]) + "\n")
			}
		}
	}
	return sb.String(), nil
}
