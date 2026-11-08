# Quickstart

This tutorial walks through creating and validating a cloud-native microservice project using CAKD Platform.

## 1. Build the Binaries

```bash
go build -o bin/cakd-cli ./cmd/cakd-cli
go build -o bin/cakd-agent ./cmd/cakd-agent
go build -o bin/cakd-mcp ./cmd/cakd-mcp
```

## 2. Option A: Use the Interactive TUI Wizard (No YAML Required)

Launch the interactive terminal wizard to generate `platform.yaml` and optionally run project creation:

```bash
./bin/cakd-cli ui
# or
./bin/cakd-cli create -i
```

## 3. Option B: Write a Minimal `platform.yaml`

Create `platform.yaml` using the canonical `cakd.io/v1alpha1` schema:

```yaml
apiVersion: cakd.io/v1alpha1
kind: Project
metadata:
  name: order-system
  owner: my-org
spec:
  cluster:
    provider: local
  backing:
    - name: orders-db
      type: postgresql
  services:
    - name: order-api
      language: java-spring-boot
      uses:
        - orders-db
```

## 4. Validate & Scaffold Locally

```bash
# Validate syntax, enums, defaults, and dependency DAG
./bin/cakd-cli validate -f platform.yaml

# Scaffold files locally without provisioning cloud resources
./bin/cakd-cli create -f platform.yaml --scaffold-only --force
```

## 5. Full Bootstrap (GitHub + ArgoCD)

Configure your credentials and run the full pipeline:

```bash
./bin/cakd-cli auth set --github-token ghp_xxx
./bin/cakd-cli init --all
./bin/cakd-cli create -f platform.yaml --force
```
