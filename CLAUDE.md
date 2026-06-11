# CAKD Platform - AI Assistant Context

## 1. Project Overview
CAKD Platform (Cloud-Agnostic Kubernetes Developer Platform) is a developer platform designed to bootstrap cloud-native microservices for small teams. It follows "Convention over Configuration", GitOps workflows, and includes a real-time agent for log and metrics diagnostic reporting.

## 2. Architecture
- **CLI Commands**:
  - `cakd` (`cmd/cakd/`): Main CLI tool for initializing, validating, creating, and diagnosing platforms.
  - `cakd-agent` (`cmd/cakd-agent/`): Background agent listening for Alertmanager webhooks to forward diagnostics.
- **Pipeline Pattern**: Coordinates the bootstrap flow in sequential steps (`internal/pipeline/`).
- **Scaffold**: Renders template engines (`internal/scaffold/`) and applies language frameworks (e.g. Spring Boot under `internal/provider/app/`).
- **IaC Engine**: `internal/iac/` and `internal/iac/terraform/` (Terraform Bridge) managing infrastructure provision.
- **Providers (`internal/provider/`)**:
  - CD (`provider/cd/argocd/`)
  - Version Control (`provider/version_control/github/`)
  - Notify (`provider/notify/discord/`)
  - LLM (`provider/llm/gemini/`)
  - Monitoring (`provider/monitoring/prometheus/`)
  - Logging (`provider/logging/loki/`)
- **Cluster Bootstrap**: Initializing kubernetes environments and installing agent workloads (`internal/cluster/`).
- **Agent Server**: Handles alert webhooks, queries loki/prometheus, and diagnoses failures using LLMs (`internal/agent/`).

## 3. Tech Stack
- Go 1.24.0 (strictly locked to ensure compatibility with CI)
- Cobra (CLI)
- Terraform (IaC)
- Helm & ArgoCD (GitOps & manifest deployments)
- Spring Boot (target application frameworks)
- Astro Starlight & TS (Documentation hub)

## 4. Coding & Documentation Conventions
- **Formatting**: Standard Go formatting (`gofmt`, `goimports` via `task fmt`).
- **Standardized Comments**: Follow Go Doc / CNCF commenting guidelines:
  - Comments on exported types, functions, and variables must start with their name (e.g. `// Engine defines...`).
  - Do not use `@author` tags.
  - Remove all scattered inline comments inside function bodies to keep code clean.
- **Error Handling**: Wrap errors with context (`fmt.Errorf("context: %w", err)`).
- **Linters**: Run `task lint` which executes `golangci-lint` (using configuration in `.golangci.yml`).
- **Commits**: Strictly follow Conventional Commits format (e.g., `feat: ...`, `fix: ...`, `docs: ...`).

## 5. Automation & Tooling
- **Taskfile (`Taskfile.yml`)**:
  - `task build` - Compiles the CLI binary into `bin/cakd`.
  - `task test` - Runs Go unit tests with race detection and generates coverage reports.
  - `task fmt` - Formats all Go files and orders imports.
  - `task vet` - Checks Go source code for structural errors.
  - `task lint` - Runs local `golangci-lint`.
  - `task ci` - Sequentially runs format, vet, lint, test, and build.
  - `task docs:generate` - Updates the Astro Starlight documentation workspace using `docs/scripts/auto-doc-agent.ts`.
- **Semantic Release**: Handles automated versioning, CHANGELOG generation, and GitHub Releases.
- **Trivy & Gitleaks**: Scanners for dependencies, secrets, and IaC vulnerability detection.

