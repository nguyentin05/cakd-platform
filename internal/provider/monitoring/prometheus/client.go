package prometheus

import (
	"encoding/json"
	"fmt"
	"os/exec"
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

// Fetch queries the Prometheus service inside the cluster (via kubectl proxy) to retrieve container status
// metrics (such as restarts) for all pods running inside the specified namespace.
func (c *Client) Fetch(namespace string) (string, error) {
	query := fmt.Sprintf(`kube_pod_container_status_restarts_total{namespace="%s"}`, namespace)
	url := fmt.Sprintf("/api/v1/namespaces/monitoring/services/prometheus-k8s:9090/proxy/api/v1/query?query=%s", query)

	cmd := exec.Command("kubectl", "get", "--raw", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query prometheus: %s", string(output))
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
