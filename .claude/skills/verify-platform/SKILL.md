---
name: verify-platform
description: Use before claiming any task, refactor phase, or PR is complete in CAKD Platform. Runs architecture boundary checks, unit/race tests, schema generation, linting, and binary smoke tests.
---

# Platform Verification Runbook

Before marking any implementation task complete or committing a milestone, run these verification checks in order and confirm their output:

## 1. Architecture Boundary Check
Ensure `domain` and `usecase` have zero forbidden imports and no core/infra package reads `os.Stdin`:
```bash
GOTOOLCHAIN=auto go test ./internal/platform/... -run TestArchitecture -v
```

## 2. Unit & Race Detector Suite
Run all unit tests with the race detector enabled:
```bash
GOTOOLCHAIN=auto go test ./... -race
```

## 3. Static Analysis & Linting
```bash
GOTOOLCHAIN=auto go vet ./...
```
*(If `golangci-lint` is required for CI readiness, run `task lint`.)*

## 4. Schema Synchronization
Regenerate the JSON Schema and ensure no uncommitted drift exists when domain structs change:
```bash
GOTOOLCHAIN=auto go run ./cmd/cakd-tools/schema-gen > schemas/platform.schema.json
```

## 5. Multi-Binary Build & Surface Smoke Tests
Compile all main binaries and verify CLI and MCP entrypoints respond cleanly:
```bash
GOTOOLCHAIN=auto go build -o bin/cakd-cli ./cmd/cakd-cli
GOTOOLCHAIN=auto go build -o bin/cakd-agent ./cmd/cakd-agent
GOTOOLCHAIN=auto go build -o bin/cakd-mcp ./cmd/cakd-mcp

./bin/cakd-cli validate -f platform.yaml
./bin/cakd-cli validate -f platform.yaml --json
printf '{"jsonrpc":"2.0","id":1,"method":"tools/list"}\n' | ./bin/cakd-cli mcp
```
