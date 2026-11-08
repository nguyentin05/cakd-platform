# CAKD Platform — AI Assistant Working Guide

CAKD Platform (Cloud-Agnostic Kubernetes Developer Platform) bootstraps cloud-native microservices, GitOps CI/CD pipelines, Kubernetes observability stacks, and AI-powered root-cause diagnostics from a single declarative configuration (`platform.yaml`) or interactive/agentic workflows.

---

## 1. Read by Task (Rules & Skills in `.claude/`)

Always read the relevant rules and skills in [`.claude/`](./.claude/) before modifying code:

| Task involves | Applicable Rule (`.claude/rules/`) | Practical Skill (`.claude/skills/`) |
| :--- | :--- | :--- |
| Architecture, package imports, layer boundaries, I/O | [01-hexagonal-boundaries.md](.claude/rules/01-hexagonal-boundaries.md) | [verify-platform](.claude/skills/verify-platform/SKILL.md) |
| `platform.yaml` schema, defaults, legacy migration, Terraform state | [02-schema-and-state.md](.claude/rules/02-schema-and-state.md) | [verify-platform](.claude/skills/verify-platform/SKILL.md) |
| Adding a new provider plugin (Cloud, VCS, CD, App, DB, LLM, Telemetry, Notify) | [01-hexagonal-boundaries.md](.claude/rules/01-hexagonal-boundaries.md), [03-go127-and-coding.md](.claude/rules/03-go127-and-coding.md) | [add-provider-plugin](.claude/skills/add-provider-plugin/SKILL.md) |
| Adding/modifying a Use Case across CLI (`--json`), TUI, and MCP Server | [01-hexagonal-boundaries.md](.claude/rules/01-hexagonal-boundaries.md) | [add-surface-capability](.claude/skills/add-surface-capability/SKILL.md) |
| Go 1.27 style, CNCF comments, error codes, secret redaction, commits | [03-go127-and-coding.md](.claude/rules/03-go127-and-coding.md) | [verify-platform](.claude/skills/verify-platform/SKILL.md) |

---

## 2. Hexagonal Architecture & Repository Map

Dependency direction is strictly inward:

```text
domain  <--  usecase (and usecase/ports)  <--  adapter (Driving) / infrastructure (Driven)  <--  composition  <--  cmd/*
```

- **`cmd/cakd-cli/`**: Developer entrypoint (CLI with `--json`, interactive TUI wizard via `cakd-cli ui`, and stdio MCP via `cakd-cli mcp`).
- **`cmd/cakd-agent/`**: In-cluster daemon listening for Alertmanager webhooks, querying Prometheus/Loki, running Gemini RCA, and dispatching Discord alerts.
- **`cmd/cakd-mcp/`**: Standalone MCP Server supporting both Stdio and Streamable HTTP (`POST /mcp` with Bearer token auth).
- **`cmd/cakd-tools/`**: Internal maintainer utilities (`registry-fetch`, `schema-gen`).
- **`internal/domain/`**: Pure Go entities (`project`, `catalog`, `diagnostic`, `errcode`). Zero external dependencies.
- **`internal/platform/`**: Technical primitives (`buildinfo`, `progress`, `logging`, `redact`) and AST architecture guard (`architecture_test.go`).
- **`internal/usecase/` & `internal/usecase/ports/`**: Application workflows (`validate`, `initcluster`, `createproject`, `observe`, `alertdispatch`, `auth`, `agenttools`, `schema`) and Outbound Port interfaces.
- **`internal/adapter/`**: Driving adapters (`cli`, `tui`, `mcpserver`, `webhook`).
- **`internal/infrastructure/`**: Driven adapters (`configfile`, `credstore`, `statestore`, `terraform`, `k8s`, `scaffold`, `providers/*`).
- **`internal/composition/`**: Shared Dependency Injection container and provider registry wiring.

---

## 3. Core Commands & Verification

- `GOTOOLCHAIN=auto go test ./... -race` — Run unit tests and AST architecture guard (`TestArchitecture`).
- `task build` — Compile binaries into `bin/`.
- `task lint` — Run `golangci-lint`.
- `task fmt` / `task vet` — Format and vet Go source code.
