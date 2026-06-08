---
title: How to Create a New CAKD Platform Project
description: This guide explains how to create a new CAKD Platform project, including generating code, setting up a GitHub repository, and registering with ArgoCD.
---

## Overview

The `cakd create` command bootstraps a new CAKD Platform `project` from a `platform.yaml` configuration. It automates code generation, GitHub repository creation, initial code push, and ArgoCD application registration.

## Prerequisites

- The `cakd` CLI binary installed and available in your PATH.
- A `platform.yaml` configuration file for your `project`.
- A GitHub Personal Access Token with `repo` scope, set as the `GITHUB_TOKEN` environment variable.
- ArgoCD installed and configured in your Kubernetes cluster (e.g., Minikube).

## Steps

### 1. Prepare your `platform.yaml` configuration

The `cakd create` command requires a `platform.yaml` file to define your `project`'s metadata and specifications. Ensure this file is correctly configured for your desired `project` name, language, and other settings.

```yaml
# Example platform.yaml (adjust as needed)
metadata:
  name: my-new-service
spec:
  language: java-spring-boot # or other supported languages
  # ... other specifications
```

:::tip
The `platform.yaml` file must be present in the directory where you run `cakd create`, or you must specify its path using the `--config` flag.
:::

### 2. Set your GitHub Token

The `Terraform Bridge` component requires a GitHub Personal Access Token to create the new repository and push the initial code. Set this token as an environment variable before running the `cakd create` command.

```bash
export GITHUB_TOKEN="ghp_YOUR_GITHUB_TOKEN"
```

:::caution
Ensure your GitHub token has the necessary permissions (e.g., `repo` scope) to create and push to repositories.
:::

### 3. Create the CAKD Platform project

Run the `cakd create` command, specifying your `platform.yaml` file. This command performs several operations:
*   Generates the `project` code, potentially downloading a base project (e.g., Spring Boot) and applying CAKD templates (Dockerfile, Helm, CI, ArgoCD).
*   Uses the `Terraform Bridge` to create a new GitHub repository.
*   Pushes the generated code to the newly created GitHub repository.
*   Registers the `project` as an ArgoCD application in your Kubernetes cluster.

```bash
cakd create --config platform.yaml
```

If a `project` directory already exists for the specified name, the command will fail unless you use the `--force` flag to overwrite it.

```bash
cakd create --config platform.yaml --force
```

:::danger
Using `--force` will delete any existing `project` directory and its contents before recreating it. Use with caution.
:::

## Verify it works

After the `cakd create` command completes successfully, it prints the repository URL and the local `project` directory path.

```bash
# No specific command needed, output is from cakd create
# Example output:
# cakd create --config platform.yaml
# ...
# ═══════════════════════════════════════════════
# Project created successfully!
# Repository: https://github.com/your-org/my-new-service
# Local code: ./out/my-new-service
# Deployment: Managed by ArgoCD (GitOps)
# ═══════════════════════════════════════════════
```

:::tip Expected result
The `cakd create` command completes without errors, displaying the "Project created successfully!" banner, the GitHub repository URL, and the local `project` directory path. You should also be able to access the new repository on GitHub and see the application registered in ArgoCD.
:::

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| `directory <project-name> already exists. Use --force to overwrite` | A `project` directory with the specified name already exists in the `out` folder. | Either choose a different `project` name in your `platform.yaml` or run `cakd create` with the `--force` flag to overwrite the existing directory. |
| `terraform failed: <error>` | The `Terraform Bridge` failed to create the GitHub repository. This could be due to invalid GitHub token permissions, network issues, or a repository with the same name already existing. | Verify your `GITHUB_TOKEN` has `repo` scope. Check GitHub for existing repositories with the same name. Review Terraform logs for specific errors. |
| `git operations failed: <error>` | The `cakd` CLI failed to initialize Git or push the code to the new GitHub repository. This often indicates an issue with the `GITHUB_TOKEN` or network connectivity. | Ensure your `GITHUB_TOKEN` is correctly set and has push permissions. Check network connectivity to GitHub. The `cakd` CLI attempts to roll back Terraform resources if this step fails. |
| `ArgoCD registration failed: <error>` | The `cakd` CLI could not register the application with ArgoCD. | Verify your Kubernetes cluster (e.g., Minikube) is running. Confirm ArgoCD is installed and accessible in your cluster. Check ArgoCD logs for more details. |