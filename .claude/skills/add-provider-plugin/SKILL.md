---
name: add-provider-plugin
description: Use when adding a new Driven Adapter (provider plugin) for Cloud Infrastructure, VCS, CI/CD, App Framework, Backing DB, LLM, Monitoring, Logging, or Notification in CAKD Platform.
---

# Adding a Provider Plugin (Driven Adapter)

Follow this 5-step workflow to add a new third-party provider plugin without violating Hexagonal boundaries:

## Step 1: Register Metadata in Domain Catalog (`internal/domain/catalog`)
1. Add the provider's identifier constant (e.g., `"gitlab"`, `"eks"`, `"go-gin"`, `"redis"`) to the appropriate enum list in `internal/domain/catalog/catalog.go`.
2. If the provider requires default versions or sizing (e.g., default language version or database version), add it to `DefaultCatalog()`.
3. Do NOT import any adapter or SDK package inside `internal/domain/catalog`.

## Step 2: Check or Extend Outbound Port (`internal/usecase/ports`)
1. Verify which interface in `internal/usecase/ports/ports.go` represents the capability:
   - App Scaffolding: `ports.AppScaffolder`
   - Cloud Host Provisioning: `ports.CloudProvisioner`
   - Kubernetes Bootstrap / Addons: `ports.ClusterBootstrapper`
   - Repository Provisioning: `ports.RepoProvisioner`
   - Version Control: `ports.VersionControl`
   - Continuous Deployment: `ports.ContinuousDeployment`
   - Notification: `ports.Notifier`
   - Observability: `ports.MetricsFetcher`, `ports.LogFetcher`, `ports.DiagnosticAI`
2. Reuse the existing port interface whenever possible. If extending a port, use domain types only.

## Step 3: Implement the Driven Adapter (`internal/infrastructure/...`)
1. Create a focused package under `internal/infrastructure/providers/<name>/` (or `internal/infrastructure/terraform/modules/<name>/` for Terraform modules).
2. Implement the port interface from `internal/usecase/ports`.
3. Rules for implementation:
   - Accept `context.Context` as the first parameter on all I/O methods.
   - Emit progress updates via `progress.Emit(ctx, ...)` instead of `fmt.Println`.
   - Never read `os.Stdin` or mutate global environment variables.
   - Wrap errors with `errcode.Wrap(...)` or `fmt.Errorf("...: %w", err)`.

## Step 4: Wire into Composition Root (`internal/composition`)
1. Register the provider factory in `internal/composition/container.go` so the use cases can resolve it by name from `project.Spec.Providers` or `project.Spec.Cluster.Provider`.
2. If the provider requires credentials in `~/.cakd/credentials.yaml`, add the non-interactive getter/setter to `internal/infrastructure/credstore` and interactive login prompt to `internal/adapter/cli/auth.go` and `internal/adapter/tui`.

## Step 5: Test & Update Schema
1. Add unit tests for the new adapter in `internal/infrastructure/providers/<name>/<name>_test.go`.
2. Run `go run ./cmd/cakd-tools/schema-gen > schemas/platform.schema.json` so IDE autocomplete and MCP schemas include the new enum value.
3. Run `go test ./... -race` and `go test ./internal/platform/... -run TestArchitecture`.
