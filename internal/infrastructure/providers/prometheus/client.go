package prometheus

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

// Client implements ports.MetricsCollector for Prometheus (in-cluster HTTP or kubectl raw proxy).
type Client struct {
	httpClient *http.Client
}

// NewClient creates a Prometheus MetricsCollector Client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type queryResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// CollectMetricsSummary fetches container restart metrics for namespace from Prometheus.
func (c *Client) CollectMetricsSummary(ctx context.Context, namespace string) (string, error) {
	query := fmt.Sprintf(`kube_pod_container_status_restarts_total{namespace="%s"}`, namespace)
	encoded := url.QueryEscape(query)

	var raw []byte
	var err error
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		apiURL := fmt.Sprintf("http://prometheus-k8s.monitoring.svc.cluster.local:9090/api/v1/query?query=%s", encoded)
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if reqErr != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "build prometheus request", reqErr)
		}
		resp, doErr := c.httpClient.Do(req)
		if doErr != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "query prometheus in-cluster", doErr)
		}
		defer resp.Body.Close()
		raw, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "read prometheus response", err)
		}
		if resp.StatusCode != http.StatusOK {
			return "", errcode.New(errcode.CodeExternal, fmt.Sprintf("prometheus status %d: %s", resp.StatusCode, string(raw)))
		}
	} else {
		path := fmt.Sprintf("/api/v1/namespaces/monitoring/services/prometheus-k8s:9090/proxy/api/v1/query?query=%s", encoded)
		//nolint:gosec
		cmd := exec.CommandContext(ctx, "kubectl", "get", "--raw", path)
		raw, err = cmd.CombinedOutput()
		if err != nil {
			return "", errcode.Wrap(errcode.CodeExternal, "query prometheus via kubectl: "+strings.TrimSpace(string(raw)), err)
		}
	}

	var parsed queryResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", errcode.Wrap(errcode.CodeExternal, "decode prometheus response", err)
	}

	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "Metrics for namespace %s:\n", namespace)
	for _, r := range parsed.Data.Result {
		pod := r.Metric["pod"]
		container := r.Metric["container"]
		if len(r.Value) == 2 {
			restarts, _ := r.Value[1].(string)
			_, _ = fmt.Fprintf(&sb, "- Pod: %s, Container: %s, Restarts: %s\n", pod, container, restarts)
		}
	}
	return sb.String(), nil
}
