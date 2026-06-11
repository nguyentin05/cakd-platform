Purpose

This file gives Copilot sessions focused, actionable guidance for working in this repository: build/test/lint commands, the high-level architecture, and repository-specific conventions found in README.md, CONTRIBUTING.md and CLAUDE.md.

Build, test, and lint commands

- Taskfile (project task runner):
  - List tasks: task --list
  - Full build: task build  (compiles cmd/cakd and cmd/cakd-agent into bin/)
  - Full test suite: task test  (CI uses: go test ./... -v -race -coverprofile=coverage.out)
  - Lint: task lint  (alias for golangci-lint run)
  - Full CI locally: task ci  (runs format → vet → lint → test → build → trivy)

- Direct Go commands (useful for single-test runs or debugging):
  - Run all tests (local / same as CI): go test ./... -v -race
  - Run a single package: go test ./internal/config -v
  - Run a single test in a package: go test ./internal/config -run '^TestParse$' -v

- golangci-lint:
  - Run lint for all code: golangci-lint run ./...
  - Config file: .golangci.yml (project enforces specific linters/formatters)

CI snippets

- Tests (GitHub Actions): go test ./... -v -race -coverprofile=coverage.out
- Lint (GitHub Actions): golangci-lint action (config: .golangci.yml)

High-level architecture (big picture)

- CLI entrypoints (cmd/):
  - cmd/cakd — CLI tool for creating and validating platforms
  - cmd/cakd-agent — daemon for Alertmanager diagnostics reporting
- internal/ contains main subsystems:
  - internal/config — platform.yaml schema parsing and validation
  - internal/registry — single source-of-truth for supported providers and defaults
  - internal/provider — provider interfaces and concrete clients (VCS github, notify discord, CD argocd, monitoring prometheus, logging loki, LLM gemini)
  - internal/pipeline — orchestrates the bootstrap step flow (Scaffold, Infra, Notify, VersionControl, Deploy)
  - internal/scaffold — template engine (go:embed) and templates for target frameworks (Spring Boot)
  - internal/iac — Terraform bridge and IaC orchestration
  - internal/cluster — minikube / cluster bootstrap and agent workload installer
  - internal/agent — alert receiver and LLM diagnostic engine

Key repository conventions (essential for Copilot suggestions)

- Registry as source of truth:
  - internal/registry/ holds supported values and defaults. Add new providers or defaults here so the parser and docs stay consistent.

- Provider pattern & factory:
  - Providers implement interfaces (see internal/provider) and are instantiated via the provider package.

- Template delimiters:
  - All Go templates use [[ ]] (NOT {{ }}) to avoid conflicts with Helm and GitHub Actions. Generated templates rely on this delimiter convention.

- Secrets and credentials:
  - platform.yaml must not contain credentials. Resolution order: environment variables → config file → interactive prompt.

- Commit messages & release:
  - Conventional Commits are enforced. Use types like feat:, fix:, docs:, test:, chore:, refactor:.

- Linting & tests:
  - golangci-lint config: .golangci.yml. Don't suppress entire files; prefer inline //nolint:reason with justification.
  - All Go doc comments must start with the name of the exported identifier, be descriptive, contain no author tags, and keep function bodies free of inline comments.

Files to consult for deeper context

- README.md — quickstart, AI observability examples, task runner usage
- CONTRIBUTING.md — development workflow, task list, registry conventions, CI task breakdown
- CLAUDE.md — AI assistant context and high-level project summary
- .github/workflows/*.yml — CI commands and expectations
- internal/registry/* and internal/provider/* — canonical places to add providers and defaults

Done