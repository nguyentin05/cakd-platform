---
trigger: always_on
description: Core Hexagonal Architecture boundaries and engineering guidelines for CAKD Platform.
---

# CAKD Platform Architecture & Guidelines

## 1. Hexagonal Architecture & Inward Dependencies

All packages inside `internal/` follow strict inward dependency:

```text
domain  <--  usecase (and usecase/ports)  <--  adapter (Driving) / infrastructure (Driven)  <--  composition  <--  cmd/*
```

- **Domain (`internal/domain/`)**: Pure business models (`project`, `catalog`, `diagnostic`) and canonical error codes (`errcode`). Zero external dependencies.
- **Use Cases & Ports (`internal/usecase/`)**: Application workflows and outbound port interfaces. Decoupled from specific cloud, VCS, or notification providers.
- **Driving Adapters (`internal/adapter/`)**: Inbound interfaces (CLI with `--json`, interactive TUI, MCP Server, Webhook daemon).
- **Driven Infrastructure (`internal/infrastructure/`)**: Concrete implementations of ports (Terraform, K8s/Helm bootstrapper, provider plugins).
- **Composition (`internal/composition/`)**: Pure constructor dependency injection root.

## 2. Core Invariants

1. **I/O & Process Isolation**:
   - Packages in `internal/domain`, `internal/usecase`, and `internal/infrastructure` must NOT call `fmt.Print*` directly to `os.Stdout`/`os.Stderr` for user-facing progress. Progress events must be emitted via `progress.Emit(ctx, ...)`.
   - Never read `os.Stdin` or call `term.ReadPassword` in core/infrastructure layers; interactive prompts belong strictly in CLI/TUI driving adapters.
   - Do NOT mutate process-wide environment variables (`os.Setenv`). Pass parameters explicitly.

2. **State Directory Isolation**:
   - Cluster provisioning state resides under `out/<project>/.platform/terraform/cluster/`.
   - VCS repository provisioning state resides under `out/<project>/.platform/terraform/vcs/`.
   - Inter-stage metadata is shared via `ports.ClusterStateStore`.

3. **Security & Idiomatic Go**:
   - Redact all sensitive tokens, keys, and webhook URLs using `redact.Secret(...)`.
   - Prefer standard library packages and wrap errors contextually using `errcode.Wrap(...)`.

## 3. Architecture Guard & Verification

Always verify boundaries before committing:

```bash
GOTOOLCHAIN=auto go test ./internal/platform/... -run TestArchitecture -v
GOTOOLCHAIN=auto go test ./... -race
```
