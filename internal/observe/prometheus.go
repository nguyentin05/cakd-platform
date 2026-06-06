package observe

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type PrometheusClient struct{}

func NewPrometheusClient() *PrometheusClient {
	return &PrometheusClient{}
}

type PrometheusResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (c *PrometheusClient) Fetch(namespace string) (string, error) {
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
