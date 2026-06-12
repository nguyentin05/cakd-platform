package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

// Client implements the [monitoring.MetricsFetcher] interface for Prometheus.
// It queries container metrics from the Prometheus server running inside the Kubernetes cluster.
type Client struct{}

// NewClient initializes and returns a new Prometheus Client.
func NewClient() *Client {
	return &Client{}
}

// PrometheusResponse represents the structured query response format returned by the Prometheus query API.
type PrometheusResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Fetch queries the Prometheus service inside the cluster (via direct HTTP if in-cluster, or via kubectl proxy if local)
// to retrieve container status metrics (such as restarts) for all pods running inside the specified namespace.
func (c *Client) Fetch(namespace string) (string, error) {
	query := fmt.Sprintf(`kube_pod_container_status_restarts_total{namespace="%s"}`, namespace)
	encodedQuery := url.QueryEscape(query)

	var output []byte
	var err error

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// In-cluster: Direct HTTP query to the Prometheus service
		apiURL := fmt.Sprintf("http://prometheus-k8s.monitoring.svc.cluster.local:9090/api/v1/query?query=%s", encodedQuery)
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		resp, httpErr := client.Get(apiURL)
		if httpErr != nil {
			return "", fmt.Errorf("failed to query prometheus directly in-cluster: %w", httpErr)
		}
		defer resp.Body.Close()

		output, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read prometheus direct response: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(output))
		}
	} else {
		// Out-of-cluster (CLI local fallback): Use kubectl get --raw
		path := fmt.Sprintf("/api/v1/namespaces/monitoring/services/prometheus-k8s:9090/proxy/api/v1/query?query=%s", encodedQuery)
		cmd := exec.Command("kubectl", "get", "--raw", path)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to query prometheus via kubectl: %s (error: %v)", string(output), err)
		}
	}

	var promResp PrometheusResponse
	if err := json.Unmarshal(output, &promResp); err != nil {
		return "", fmt.Errorf("failed to parse prometheus response: %w", err)
	}

	var result string
	result += fmt.Sprintf("Metrics for namespace %s:\n", namespace)
	for _, res := range promResp.Data.Result {
		pod := res.Metric["pod"]
		container := res.Metric["container"]
		if len(res.Value) == 2 {
			restarts := res.Value[1].(string)
			result += fmt.Sprintf("- Pod: %s, Container: %s, Restarts: %s\n", pod, container, restarts)
		}
	}

	return result, nil
}
