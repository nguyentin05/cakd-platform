---
title: Internal Architecture
description: How CAKD Platform works internally — components, layers, and execution flow.
---

## Overview

CAKD Platform is a CLI that `bootstrap`s cloud-native projects from a single `platform.yaml`. It parses the config, renders embedded templates, provisions external resources (GitHub repository via Terraform), pushes generated code, and registers the application with ArgoCD for GitOps deployment.

## Architecture Layers

### Configuration

**Responsibility:** Parse `platform.yaml`, validate structure and logic, and apply defaults.

**Components:** `internal/config` (parse, validate, defaults)

This layer enforces a three-phase pipeline: structure validation, defaults injection, and logic/dependency validation. It produces a hydrated `PlatformConfig` passed to downstream components.

### Orchestration / Pipeline

**Responsibility:** Coordinate the bootstrap steps and error handling.

**Components:** `internal/pipeline` (Execute, Run, Step implementations)

The pipeline assembles ordered steps (Scaffold, Infra, Notify, VersionControl, Deploy). Each step implements `Step` and optional `OptionalStep` for conditional execution.

### Template Engine

**Responsibility:** Render project files from embedded Go templates.

**Components:** `internal/scaffold` (Engine, renderTemplate)

Templates are embedded via `go:embed` and rendered with `[[ ]]` delimiters. Engine generates global and service-specific artifacts (Dockerfile, Helm charts, CI workflows) into the project output directory.

### Infrastructure Provisioning (Terraform Bridge)

**Responsibility:** Provision external resources required by the project, e.g., GitHub repositories.

**Components:** `internal/iac/terraform` (Bridge)

The Terraform Bridge copies embedded modules, writes `terraform.tfvars.json`, runs `terraform init` and `terraform apply`, and parses outputs (e.g., repo URL) for subsequent steps.

### Version Control & GitOps

**Responsibility:** Initialize local git, commit generated files, push to remote, and register with ArgoCD.

**Components:** `internal/provider/version_control` (GitHub client), `internal/git`, `internal/provider/cd/argocd`

The VCS provider initializes a repo, commits, and pushes via authenticated clone URL. After push, the pipeline registers an ArgoCD application pointing at the repository's manifest.

### Observability & Agent

**Responsibility:** Collect metrics/logs and run AI diagnostics for alerts.

**Components:** `internal/observe` (ObserverService), `internal/agent` (cakd-agent), providers: `prometheus`, `loki`, `llm/gemini`

The Agent listens for Alertmanager webhooks, the ObserverService fetches Prometheus and Loki data, then submits a formatted prompt to Gemini and delivers results via notification providers.

## Execution Flow

### Project creation (`cakd create -f platform.yaml`)

1. Parse and validate `platform.yaml` → produce `PlatformConfig`.
2. Prepare output directory; honor `--force` to overwrite.
3. Scaffold base project (e.g., Spring Initializr) and render CAKD templates into `out/{project}`.
4. Run Terraform Bridge to create GitHub repository and retrieve outputs.
5. Initialize git, commit generated files, and push to GitHub (VCS provider).
6. Register the generated ArgoCD application manifest with ArgoCD.
7. Print summary with repository URL and local path.

### Observability flow (`cakd observe` / `cakd-agent`)

1. Receive a project name or an Alertmanager webhook.
2. Query Prometheus (Prometheus client) for pod/container metrics.
3. Query Loki (Loki client) for recent logs.
4. Build an AI prompt that includes metrics and logs.
5. Send the prompt to Gemini (LLM client) and receive a diagnosis.
6. Format and deliver the diagnosis via configured notifier (Discord).

## Component Diagram

```mermaid
graph TD
  CLI[cakd CLI] --> Pipeline[internal/pipeline]
  Pipeline --> Scaffold[internal/scaffold]
  Pipeline --> Terraform[internal/iac/terraform]
  Pipeline --> VCS[internal/provider/version_control]
  Pipeline --> CD[internal/provider/cd]
  Pipeline --> Notify[internal/provider/notify]

  Agent[cakd-agent] --> Observe[internal/observe]
  Observe --> Prom[internal/provider/monitoring/prometheus]
  Observe --> Loki[internal/provider/logging/loki]
  Observe --> LLM[internal/provider/llm/gemini]
  Observe --> Notify
```

## Key Design Decisions

- Declarative `platform.yaml` centralizes project definition and enables repeatable bootstraps.
- Modular packages provide clear separation (config, pipeline, scaffold, iac, provider implementations).
- Embedded templates (`go:embed`) and `[[ ]]` delimiters avoid conflicts with Helm/CI syntaxes.
- Terraform Bridge isolates infrastructure provisioning and returns structured outputs used by the pipeline.
