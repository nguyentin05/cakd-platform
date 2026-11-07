package agenttools_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func TestAgentToolsCatalogResourcesAndValidateTool(t *testing.T) {
	container := composition.NewContainer()
	tools := container.AgentTools

	if len(tools.ListTools()) < 6 {
		t.Fatalf("expected at least 6 MCP tools, got %d", len(tools.ListTools()))
	}
	if len(tools.ListResources()) < 2 {
		t.Fatalf("expected at least 2 MCP resources, got %d", len(tools.ListResources()))
	}

	schemaContent, _, err := tools.ReadResource(context.Background(), "cakd://schema/project")
	if err != nil {
		t.Fatalf("ReadResource schema failed: %v", err)
	}
	if !strings.Contains(schemaContent, "cakd.io/v1alpha1") {
		t.Fatalf("expected schema resource to mention cakd.io/v1alpha1")
	}

	out, err := tools.CallTool(context.Background(), "cakd_generate_spec", map[string]any{
		"name":        "payment-system",
		"owner":       "acme",
		"serviceName": "payment-api",
		"provider":    "aws-ec2",
	})
	if err != nil {
		t.Fatalf("CallTool cakd_generate_spec failed: %v", err)
	}
	if !strings.Contains(out, "payment-system") || !strings.Contains(out, "spec:") {
		t.Fatalf("unexpected generated spec:\n%s", out)
	}

	valOut, err := tools.CallTool(context.Background(), "cakd_validate_spec", map[string]any{
		"yamlContent": out,
	})
	if err != nil {
		t.Fatalf("CallTool cakd_validate_spec failed: %v", err)
	}
	if !strings.Contains(valOut, `"valid": true`) {
		t.Fatalf("expected valid=true response, got: %s", valOut)
	}
}
