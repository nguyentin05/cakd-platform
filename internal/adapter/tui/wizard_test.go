package tui_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/adapter/tui"
	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func TestWizardGeneratesCanonicalSpecFile(t *testing.T) {
	c := composition.NewContainer()
	outPath := filepath.Join(t.TempDir(), "platform.yaml")

	// Inputs:
	// 1. Project name: order-system
	// 2. Owner: acme
	// 3. Cluster provider: aws-ec2
	// 4. Region: ap-southeast-1
	// 5. Instance type: t3.medium
	// 6. Service name: order-api
	// 7. Java version: 21
	// 8. Build tool: maven-project
	// 9. Add PostgreSQL?: y
	// 10. DB name: orders-db
	// 11. Run create now?: n
	input := strings.Join([]string{
		"order-system",
		"acme",
		"aws-ec2",
		"ap-southeast-1",
		"t3.medium",
		"order-api",
		"21",
		"maven-project",
		"y",
		"orders-db",
		"n",
	}, "\n") + "\n"

	var out bytes.Buffer
	err := tui.RunWizard(context.Background(), strings.NewReader(input), &out, &tui.Services{
		ValidateUC: c.ValidateUC,
		SpecRepo:   c.SpecRepo,
		CreateUC:   c.CreateUC,
	}, tui.Options{
		OutputPath: outPath,
	})
	if err != nil {
		t.Fatalf("RunWizard failed: %v", err)
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("expected %s to be written: %v", outPath, err)
	}
	content := string(raw)
	for _, want := range []string{"apiVersion: cakd.io/v1alpha1", "name: order-system", "provider: aws-ec2", "name: orders-db", "name: order-api"} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected generated YAML to contain %q, got:\n%s", want, content)
		}
	}
}
