package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli"
	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func TestCLIVersionSchemaAndValidateJSON(t *testing.T) {
	c := composition.NewContainer()

	var out bytes.Buffer
	root := cli.NewRootCommand(c, strings.NewReader(""), &out, &out)
	root.SetArgs([]string{"version", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("version --json failed: %v", err)
	}
	if !strings.Contains(out.String(), `"version"`) {
		t.Fatalf("expected JSON version output, got: %s", out.String())
	}

	out.Reset()
	schemaPath := filepath.Join(t.TempDir(), "platform.schema.json")
	root = cli.NewRootCommand(c, strings.NewReader(""), &out, &out)
	root.SetArgs([]string{"schema", "-o", schemaPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("schema command failed: %v", err)
	}

	specPath := filepath.Join(t.TempDir(), "platform.yaml")
	yamlStr, err := c.AgentTools.CallTool(t.Context(), "cakd_generate_spec", map[string]any{
		"name":        "cli-test",
		"owner":       "acme",
		"serviceName": "api",
		"outputPath":  specPath,
	})
	if err != nil || yamlStr == "" {
		t.Fatalf("generate spec failed: %v", err)
	}

	out.Reset()
	root = cli.NewRootCommand(c, strings.NewReader(""), &out, &out)
	root.SetArgs([]string{"validate", "-f", specPath, "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("validate -f --json failed: %v", err)
	}
	if !strings.Contains(out.String(), `"valid": true`) {
		t.Fatalf("expected valid: true in JSON output, got: %s", out.String())
	}
}
