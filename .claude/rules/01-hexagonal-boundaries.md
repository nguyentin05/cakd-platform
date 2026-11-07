---
trigger: always_on
description: Strict Hexagonal Architecture (Ports & Adapters) layer boundaries and I/O isolation rules for CAKD Platform.
---

# Rule 01: Hexagonal Architecture & Layer Boundaries

## 1. Inward-Only Dependency Rule

All packages inside `internal/` MUST strictly follow the inward dependency direction:

```text
domain  <--  usecase (and usecase/ports)  <--  adapter (Driving) / infrastructure (Driven)  <--  composition  <--  cmd/*
```

| Layer | Path | Allowed Internal Imports | Forbidden Imports |
| :--- | :--- | :--- | :--- |
| **Domain** | `internal/domain/...` | Only `internal/domain/...` and Go stdlib | `usecase`, `adapter`, `infrastructure`, `platform`, `composition`, or any third-party SDK |
| **Platform** | `internal/platform/...` | Go stdlib and `internal/domain/...` | `usecase`, `adapter`, `infrastructure`, `composition` |
| **Use Case & Ports** | `internal/usecase/...` | `internal/domain/...`, `internal/platform/...`, `internal/usecase/ports` | `internal/adapter/...`, `internal/infrastructure/...`, `internal/composition/...`, or concrete third-party clients |
| **Driving Adapters** | `internal/adapter/...` | `internal/usecase/...`, `internal/domain/...`, `internal/platform/...` | `internal/infrastructure/...` (except via `composition` in entrypoint wiring) |
| **Driven Adapters** | `internal/infrastructure/...` | `internal/usecase/ports`, `internal/domain/...`, `internal/platform/...` | `internal/adapter/...`, `internal/usecase/*` (concrete use-case services) |
| **Composition** | `internal/composition/...` | All `internal/...` packages | `cmd/...` |

## 2. Strict Terminal & Process I/O Isolation

1. **No Direct Stdout/Stderr in Core**: Packages in `internal/domain`, `internal/usecase`, and `internal/infrastructure` MUST NOT call `fmt.Print`, `fmt.Println`, `fmt.Printf` to `os.Stdout`/`os.Stderr` for user-facing progress. All progress updates MUST be emitted via `progress.Emit(ctx, progress.Event{...})` (`internal/platform/progress`) so CLI, TUI, and MCP Server can render or suppress them appropriately.
2. **No Interactive Stdin in Core or Infrastructure**: Packages in `internal/domain`, `internal/usecase`, and `internal/infrastructure` MUST NEVER read `os.Stdin` or call `term.ReadPassword`. Interactive credential prompting is strictly restricted to Driving Adapters (`internal/adapter/cli` and `internal/adapter/tui`).
3. **No Process-Wide Environment Mutation**: Adapters MUST NOT call `os.Setenv("KUBECONFIG", ...)` to pass state between steps. Always pass `kubeconfigPath` explicitly through `ports` structs or set `cmd.Env` on the specific `exec.Cmd` instance.

## 3. Mandatory Architecture Guard Verification

Whenever adding or modifying imports across `internal/`, run the AST architecture test:

```bash
go test ./internal/platform/... -run TestArchitecture -v
```
