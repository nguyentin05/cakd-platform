---
title: Quickstart
description: Bootstrap your first cloud-native project with CAKD Platform and start monitoring in under 10 minutes.
sidebar:
  order: 1
---

## Prerequisites

Before you begin, ensure you have:

- [ ] [Go 1.24+](https://go.dev/dl/) installed
- [ ] [Terraform](https://developer.hashicorp.com/terraform/install) installed
- [ ] [Minikube](https://minikube.sigs.k8s.io/docs/start/) installed and running
- [ ] A [GitHub Personal Access Token](https://github.com/settings/tokens) with `repo` scope
- [ ] `GITHUB_TOKEN` environment variable set

## Step 1: Install CAKD & CAKD Agent

Install the `cakd` CLI and `cakd-agent` into your Go bin path:

```bash
go install github.com/nguyentin05/cakd-platform/cmd/cakd@latest
go install github.com/nguyentin05/cakd-platform/cmd/cakd-agent@latest
```

## Step 2: Create your `platform.yaml`

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

Run the create pipeline to generate project files, provision a GitHub repository via the Terraform Bridge, push the generated code, and register an ArgoCD application.

```bash
cakd create -f platform.yaml
```

:::tip Expected output
You will see progress logs for each step of the pipeline, followed by a summary:

```
Step 1/5: Scaffold...
Step 2/5: Infra...
Step 3/5: VersionControl...
Step 4/5: Deploy...
Step 5/5: Notify...

═══════════════════════════════════════════════
Project created successfully!
Repository: https://github.com/your-organization/my-first-project
Local code: out/my-first-project
Deployment: Managed by ArgoCD (GitOps)
═══════════════════════════════════════════════
```
:::

## Step 4: Start the Alert & Diagnostics Agent

If you want real-time diagnostics, run the agent after setting a webhook receiver (Discord):

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
cakd-agent
```

## Step 5: Verify the result

Confirm the project folder exists and ArgoCD has the application registered:

```bash
ls -F out/my-first-project/
kubectl get app my-first-project -n argocd
```

## What was created

- A local project under `out/{project-name}` populated by the template engine.
- A GitHub repository provisioned via the Terraform Bridge and pushed by the VCS provider.
- An ArgoCD Application registered in your cluster to deploy the project.

## Next Steps

- [Full CLI reference](/reference/cli/)
- [How to set up Discord alerts](/how-to-guides/setup-discord-alerts/)