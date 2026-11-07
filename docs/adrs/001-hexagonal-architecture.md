# ADR 001: Hexagonal Architecture (Ports & Adapters)

- **Status**: Accepted
- **Date**: 2026-11-07

## Context

Earlier iterations of CAKD coupled CLI terminal I/O (`os.Stdout`, `os.Stdin`, `term.ReadPassword`) and process-wide environment mutations (`os.Setenv("KUBECONFIG", ...)`) directly inside infrastructure and config packages, making non-interactive surfaces (TUI, MCP Server) unsafe.

## Decision

Adopt a strict Hexagonal Architecture (`domain <-- usecase/ports <-- adapter / infrastructure <-- composition <-- cmd/*`) guarded by an automated Go AST boundary test (`internal/platform/architecture_test.go`).

## Consequences

- Core domain and application use cases are 100% pure and reusable across CLI, TUI, MCP Server, and Webhook daemon surfaces.
- Interactive prompts are isolated strictly to `internal/adapter/cli` and `internal/adapter/tui`.
