# CAKD Platform — AI Assistant Working Guide

CAKD Platform (Cloud-Agnostic Kubernetes Developer Platform) bootstraps cloud-native microservices, GitOps CI/CD pipelines, Kubernetes observability stacks, and AI root-cause diagnostics from a single declarative manifest (`platform.yaml`) or interactive/agentic workflows.

---

## 1. Architecture at a Glance

CAKD follows **Hexagonal Architecture (Ports & Adapters)** with inward-only dependencies:

```text
domain  <--  usecase (and usecase/ports)  <--  adapter / infrastructure  <--  composition  <--  cmd/*
```

- **Domain (`internal/domain/`)**: Pure entities (`project`, `catalog`, `diagnostic`) and canonical error codes (`errcode`).
- **Use Cases & Ports (`internal/usecase/`)**: Core application workflows (`validate`, `initcluster`, `createproject`, `observe`, `alertdispatch`) and outbound port interfaces.
- **Driving Adapters (`internal/adapter/`)**: External entry surfaces — Developer CLI (`cakd-cli`), interactive TUI wizard (`ui`), MCP Server (`mcp`), and Alertmanager webhook daemon (`cakd-agent`).
- **Driven Infrastructure (`internal/infrastructure/`)**: Concrete implementations of ports — Terraform modules, K8s/Helm bootstrapper, template scaffolding, and provider clients (GitHub, ArgoCD, Discord, Gemini, Prometheus, Loki).
- **Composition (`internal/composition/`)**: Pure constructor dependency injection container.

For core boundary rules and invariants, see [`.claude/rules/01-architecture.md`](.claude/rules/01-architecture.md).

---

## 2. Core Development Commands

- `task test` or `go test ./... -race` — Run unit tests and AST architecture boundary guard (`TestArchitecture`).
- `task build` — Compile platform binaries (`bin/cakd-cli`, `bin/cakd-agent`, `bin/cakd-mcp`).
- `task lint` — Run `golangci-lint`.
- `task fmt` / `task vet` — Format and vet Go source code.
