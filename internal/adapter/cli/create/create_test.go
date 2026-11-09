package create_test

import (
	"bytes"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/create"
)

func TestNewCommandFlags(t *testing.T) {
	const defaultFalse = "false"
	f := cmdutil.NewFactoryWith(&cmdutil.Services{}, &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	cmd := create.NewCommand(f)

	if cmd.Use != "create" {
		t.Fatalf("expected command use 'create', got %q", cmd.Use)
	}

	forceFlag := cmd.Flag("force")
	if forceFlag == nil || forceFlag.DefValue != defaultFalse {
		t.Fatalf("expected --force flag")
	}

	scaffoldFlag := cmd.Flag("scaffold-only")
	if scaffoldFlag == nil || scaffoldFlag.DefValue != defaultFalse {
		t.Fatalf("expected --scaffold-only flag")
	}

	interactiveFlag := cmd.Flag("interactive")
	if interactiveFlag == nil || interactiveFlag.DefValue != defaultFalse {
		t.Fatalf("expected --interactive flag")
	}
}
