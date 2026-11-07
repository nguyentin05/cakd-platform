# Dependency Decisions

CAKD Platform keeps external Go module dependencies minimal to ensure fast builds, small binaries, and a minimal CVE surface:

- `github.com/spf13/cobra` — CLI command tree and flag parsing
- `gopkg.in/yaml.v3` — YAML parsing and serialization
- `golang.org/x/crypto` — SSH client for remote K3s bootstrap
- `golang.org/x/term` — Secure terminal password input at the CLI adapter layer
