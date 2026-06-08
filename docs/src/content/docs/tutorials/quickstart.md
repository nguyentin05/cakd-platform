---
title: Quickstart
description: Bootstrap your first cloud-native project with CAKD Platform in under 10 minutes.
sidebar:
  order: 1
---

## Prerequisites

Before you begin, ensure you have:

- [ ] [Go 1.21+](https://go.dev/dl/) installed
- [ ] [Terraform](https://developer.hashicorp.com/terraform/install) installed  
- [ ] [Minikube](https://minikube.sigs.k8s.io/docs/start/) installed and running
- [ ] A [GitHub Personal Access Token](https://github.com/settings/tokens) with `repo` scope
- [ ] `GITHUB_TOKEN` environment variable set

## Step 1: Install CAKD

Install the `cakd` CLI binary using `go install`. This command fetches the source code, compiles it, and places the executable in your Go binary path (e.g., `~/go/bin`).

```bash
go install github.com/nguyentin05/cakd-platform/cmd/cakd@latest
```

## Step 2: Create your `platform.yaml`

Create a file named `platform.yaml` with the following minimal configuration. This defines the basic properties of your new `project`.

```yaml
apiVersion: platform.cakd.dev/v1alpha1
kind: Project
metadata:
  name: my-first-project
  owner: my-team
spec:
  language: java-spring-boot
  version: 3.2.0
```

## Step 3: Bootstrap your project

Run the `cakd create` command, pointing it to your `platform.yaml` file. This command will `bootstrap` your `project` by generating code, creating a GitHub repository, pushing the code, and registering an ArgoCD application.

```bash
cakd create -f platform.yaml
```

:::tip Expected output
A successful run of `cakd create` displays progress messages for each step, followed by a summary:

```text
Starting create for project: my-first-project
Step 1/4: Generating project...
   Downloading base project from start.spring.io...
   Spring Boot base project generated (official)
   Applying CAKD templates (Dockerfile, Helm, CI, ArgoCD)...
   Project ready at: ./out/my-first-project
Step 2/4: Creating GitHub repository via Terraform...
Step 3/4: Pushing code to repository...
Step 4/4: Registering ArgoCD application...
   ArgoCD application registered successfully

═══════════════════════════════════════════════
Project created successfully!
Repository: https://github.com/my-team/my-first-project
Local code: ./out/my-first-project
Deployment: Managed by ArgoCD (GitOps)
═══════════════════════════════════════════════
```
(Note: The actual GitHub repository URL will vary based on your GitHub organization/user and project name.)
:::

## Step 4: Verify the result

Verify that the `project` directory was created locally and that the ArgoCD application is registered.

```bash
ls -F out/my-first-project/
kubectl get app my-first-project -n argocd
```

## What was created

The `cakd create` command generated the following:

-   A local `project` directory named `out/my-first-project` containing:
    -   A base Spring Boot `project` (if `java-spring-boot` was specified).
    -   CAKD-specific templates, including a `Dockerfile`, Helm charts, CI/CD pipeline configurations, and an ArgoCD application manifest. This was handled by the `template engine`.
-   A new GitHub repository (e.g., `https://github.com/my-team/my-first-project`) with the generated code pushed to it. This was handled by the `Terraform Bridge`.
-   An ArgoCD application registered in your Kubernetes cluster, configured to deploy your new `project` from the GitHub repository.

## Next Steps

- [Full CLI reference](/reference/cli/)
- [How to set up Discord alerts](/how-to-guides/setup-discord-alerts/)