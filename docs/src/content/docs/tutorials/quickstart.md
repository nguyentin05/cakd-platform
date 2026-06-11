---
title: Quickstart
description: Bootstrap your first cloud-native project with CAKD Platform in under 10 minutes.
sidebar:
  order: 1
---

## Prerequisites

Before you begin, ensure you have:

- [ ] [Go 1.24+](https://go.dev/dl/) installed
- [ ] [Terraform](https://developer.hashicorp.com/terraform/install) installed
- [ ] A Kubernetes cluster (Minikube or similar) with kubeconfig access
- [ ] A GitHub Personal Access Token with `repo` and `delete_repo` scopes
- [ ] `GITHUB_TOKEN` environment variable set

## Step 1: Install the `cakd` CLI

Install the CLI into your Go bin path:

```bash
go install github.com/nguyentin05/cakd-platform/cmd/cakd@latest
```

## Step 2: Create a minimal `platform.yaml`

Create `platform.yaml` in your working directory. This minimal example is valid against the schema shown in the repository.

```yaml
apiVersion: platform.dev/v1alpha1
kind: Project
metadata:
  name: my-first-project
  owner: my-team
services:
  - name: api
    language: java-spring-boot
    languageVersion: "21"
    projectBuild: gradle-project
    packaging: jar
```

## Step 3: Bootstrap your project

Run the create pipeline to generate project files, provision a GitHub repository via Terraform, push the generated code, and register an ArgoCD application.

```bash
cakd create -f platform.yaml
```

:::tip Expected output
You will see step progress logs (Scaffold, Infra, Notify, VersionControl, Deploy) and a final summary showing the repository URL and local output path.
:::

## Step 4: Start the agent (optional)

If you want real-time diagnostics, run the agent after setting a webhook receiver (Discord):

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
cakd-agent
```

## Step 5: Verify

Confirm the project folder exists and ArgoCD has the application registered:

```bash
ls -F out/my-first-project/
kubectl get app my-first-project -n argocd
```

## What was created

- A local project under `out/{project-name}` populated by the template engine.
- A GitHub repository provisioned via the Terraform Bridge and pushed by the VCS provider.
- An ArgoCD Application registered in your cluster to deploy the project.

## Next steps

- Read the full CLI reference: [/reference/cli/](/reference/cli/)
- Configure alerting/notifications: [/how-to-guides/setup-discord-alerts/](/how-to-guides/setup-discord-alerts/)
