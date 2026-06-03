# CAKD Platform - AI Assistant Context

## 1. Project Overview
CAKD Platform (Cloud-Agnostic Kubernetes Developer Platform) is a CLI tool designed to bootstrap cloud-native microservices for small teams. It follows "Convention over Configuration" and GitOps workflows.

## 2. Architecture
- **CLI Framework**: Uses `cobra` (`cmd/cakd`).
- **Parser**: Reads and validates `platform.yaml` (`internal/config`).
- **Template Engine**: Uses Go's `text/template` and `go:embed`. Uses `[[ ]]` delimiters instead of `{{ }}` to avoid conflicts with Helm and GitHub Actions (`internal/template`).
- **Terraform Bridge**: Invokes embedded Terraform modules to provision infrastructure (e.g., GitHub repos) (`internal/terraform`).
- **Git Automation**: Automates git init, commit, and push (`internal/git`).

## 3. Tech Stack
- Go 1.21+
- Cobra (CLI)
- Terraform (IaC)
- Helm (Kubernetes manifests)
- Spring Boot (Target application templates)

## 4. Coding Conventions
- **Go**: Follow standard Go formatting (`gofmt`, `goimports`).
- **Error Handling**: Wrap errors with context (`fmt.Errorf("do something: %w", err)`).
- **Linters**: Must pass `golangci-lint` with the strict configuration in `.golangci.yml`.
- **Commits**: Strictly follow Conventional Commits format (e.g., `feat: ...`, `fix: ...`).

## 5. Automation & Tooling
- **Taskfile**: Use `Taskfile.yml` as the task runner (e.g., `task build`, `task lint`, `task test`).
- **Semantic Release**: Handles automated versioning, CHANGELOG generation, and GitHub Releases.
- **GoReleaser**: Cross-compiles binaries and attaches them to GitHub Releases.
- **Trivy & Gitleaks**: Security scanners integrated into the CI/CD pipeline.
