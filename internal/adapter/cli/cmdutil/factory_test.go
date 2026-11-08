package cmdutil_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
)

func TestNewFactoryWithDefaultsAndWriteJSON(t *testing.T) {
	f := cmdutil.NewFactoryWith(nil, nil, nil, nil)
	if f.In != os.Stdin || f.Out != os.Stdout || f.ErrOut != os.Stderr {
		t.Fatal("nil streams must fall back to os streams")
	}

	var out bytes.Buffer
	f = cmdutil.NewFactoryWith(nil, nil, &out, nil)
	if err := f.WriteJSON(map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "{\n  \"a\": 1\n}\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
