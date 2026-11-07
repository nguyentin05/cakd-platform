---
trigger: always_on
description: Go 1.27 coding standards, CNCF documentation style, error handling, secret redaction, and commit conventions.
---

# Rule 03: Go 1.27 Coding Standards, Documentation & Security

## 1. Go 1.27 & Dependencies

- Target Go version is strictly **`go 1.27.0`** in `go.mod`.
- Prefer standard library packages (`log/slog`, `maps`, `slices`, `context`, `errors`) over third-party equivalents.
- Use constructor injection at the composition root (`internal/composition` and `cmd/*`). Do not introduce reflection-based DI frameworks or mutable package-level singletons.

## 2. CNCF / GoDoc Comment Conventions

- Every exported type, interface, function, method, and constant/variable MUST have a GoDoc comment starting with its exact symbol name (e.g., `// Service coordinates...`).
- Do NOT use `@author` tags.
- Keep function bodies clean: avoid scattered inline comments unless explaining a non-obvious security or protocol invariant.

## 3. Error Handling & Closed-Set Domain Error Codes

- Always wrap errors with contextual information using `fmt.Errorf("context: %w", err)` or `errcode.Wrap(code, "context", err)`.
- Map domain and adapter errors to the canonical codes in `internal/domain/errcode` (`CodeValidation`, `CodeAuth`, `CodeNotFound`, `CodeProvision`, `CodeExternal`) so CLI (`--json`), TUI, and MCP Server can return deterministic, machine-readable error responses.

## 4. Security & Secret Redaction

- Never log raw tokens, API keys, SSH private keys, or Discord webhook URLs. Use `redact.Secret(...)` (`internal/platform/redact`) whenever displaying credential status.
- Execute external binaries (`terraform`, `helm`, `kubectl`, `git`) using explicit argv slices (`exec.CommandContext(ctx, bin, args...)`), never via `sh -c` with unvalidated user strings.
