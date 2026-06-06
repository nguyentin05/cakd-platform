---
title: CLI Reference
description: Command line interface reference.
---

The `cakd` command-line interface (CLI) is the central tool for managing the Cloud-Agnostic Kubernetes Developer (CAKD) Platform. Use it to initialize platform infrastructure, create new projects, validate configurations, and run AI-powered diagnostic observation on your workloads.

## Global Command

### `cakd`

CAKD Platform CLI — Cloud-Agnostic Kubernetes Developer Platform.

```bash
cakd [command]
```

#### Available Commands
* [`init`](#cakd-init) — Bootstrap the Kubernetes cluster with platform infrastructure.
* [`create`](#cakd-create) — Generate a new cloud-native project.
* [`validate`](#cakd-validate) — Verify a `platform.yaml` configuration file.
* [`observe`](#cakd-observe) — Diagnose and troubleshoot projects using AI.
* [`version`](#cakd-version) — View current CLI version details.

---

## `cakd init`

Bootstrap your Kubernetes cluster with core platform infrastructure including ArgoCD, Prometheus, and Loki. 

```bash
cakd init [flags]
```

> **Note:** If no flags are specified, `cakd init` defaults to installing **all** components (`--argocd`, `--monitoring`, and `--logging`).

### Flags

| Flag | Shorthand | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--argocd` | - | boolean | `false` | Install ArgoCD. |
| `--monitoring` | - | boolean | `false` | Install Prometheus & Grafana. |
| `--logging` | - | boolean | `false` | Install Loki. |

### Examples

Initialize all cluster infrastructure components:
```bash
cakd init
```

Install only ArgoCD and Logging components:
```bash
cakd init --argocd --logging
```

---

## `cakd create`

Create a new cloud-native project defined in a configuration file.

```bash
cakd create [flags]
```

### Flags

| Flag | Shorthand | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--file` | `-f` | string | `"platform.yaml"` | Path to the `platform.yaml` configuration file. |
| `--force` | - | boolean | `false` | Force overwrite the existing project directory if it already exists. |

### Examples

Generate a project using the default config file location:
```bash
cakd create
```

Force-overwrite an existing project using a custom configuration file:
```bash
cakd create --file ./deployments/custom-platform.yaml --force
```

---

## `cakd validate`

Validate the syntax and structure of a `platform.yaml` configuration file without deploying any resources.

```bash
cakd validate [flags]
```

### Flags

| Flag | Shorthand | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--file` | `-f` | string | `"platform.yaml"` | Path to the `platform.yaml` file to validate. |

### Examples

Validate the `platform.yaml` in the current working directory:
```bash
cakd validate
```

Validate a specific configuration file:
```bash
cakd validate -f ./configs/my-app.yaml
```

---

## `cakd observe`

Diagnose, troubleshoot, and observe a running project using AI integration. This command gathers telemetry metrics and log data to output a holistic diagnosis of your project.

```bash
cakd observe <project-name> [flags]
```

### Arguments

| Argument | Type | Description |
| :--- | :--- | :--- |
| `<project-name>` | string (Required) | The exact name of the project/workload to diagnose. |

### Environment Variables

This command requires connection to the Google Gemini API. You must configure the following variable in your environment:

| Variable Name | Description |
| :--- | :--- |
| `GEMINI_API_KEY` | Your Gemini AI API credential. |

### Examples

Export your API Key and run diagnosis on a project named `payment-gateway`:
```bash
export GEMINI_API_KEY="your-gemini-api-key-here"
cakd observe payment-gateway
```

---

## `cakd version`

Print the version information of the CAKD Platform CLI.

```bash
cakd version
```