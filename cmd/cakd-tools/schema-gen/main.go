package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/usecase/schema"
)

func main() {
	var outPath string
	flag.StringVar(&outPath, "o", "schemas/platform.schema.json", "Output path for generated JSON Schema")
	flag.Parse()

	data, err := schema.GenerateJSONSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate JSON Schema: %v\n", err)
		os.Exit(1)
	}
	if dir := filepath.Dir(outPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
	if err := os.WriteFile(outPath, append(data, '\n'), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write JSON Schema to %s: %v\n", outPath, err)
		os.Exit(1)
	}
	fmt.Printf("Generated JSON Schema at %s\n", outPath)
}
