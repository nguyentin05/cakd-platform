---
title: cakd init
description: Bootstrap the Kubernetes cluster with Platform Infrastructure (ArgoCD, Prometheus, Loki).
sidebar:
  order: 300
---

## Overview

The `cakd init` command bootstraps your Kubernetes cluster by installing essential platform infrastructure components. This command is typically run after you have provisioned a Kubernetes cluster and before deploying any `cakd` projects. It deploys core services like ArgoCD for GitOps, Prometheus and Grafana for monitoring, and Loki for logging, preparing your cluster for cloud-native workloads.

## Usage

```bash
cakd init [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
|`--argocd`||`bool`|`false`|Install ArgoCD|
|`--monitoring`||`bool`|`false`|Install Prometheus & Grafana|
|`--logging`||`bool`|`false`|Install Loki|

## How It Works

1.  The command first checks if any specific infrastructure components (ArgoCD, Monitoring, Logging) are requested via flags. If no flags are provided, all three components are selected for installation by default.
2.  It then constructs a list of installation steps based on the selected components.
3.  For each selected component, a series of `helm` commands are prepared:
    *   **ArgoCD:** Adds steps to add the `argo` Helm repository, update it, and then install `argo-cd` using `helm upgrade --install` into the `argocd` namespace.
    *   **Monitoring (Prometheus & Grafana):** Adds steps to add the `prometheus-community` Helm repository, update it, and then install `kube-prometheus-stack` using `helm upgrade --install` into the `monitoring` namespace.
    *   **Logging (Loki):** Adds steps to add the `grafana` Helm repository, update it, and then install `loki-stack` using `helm upgrade --install` into the `monitoring` namespace.
4.  If no components are selected (e.g., if flags were explicitly set to `false` for all, or if the default was overridden and nothing was selected), the command exits, indicating no installation.
5.  The command iterates through the prepared installation steps. For each step, it executes the associated `helm` commands sequentially.
6.  Standard output and error from the `helm` commands are streamed to the console. If any `helm` command fails, the `cakd init` command terminates with an error.
7.  Upon successful completion of all steps, a success message is displayed, along with confirmation of the installed components and their namespaces.

## Examples

### Basic usage
Run `cakd init` to install all default platform infrastructure components (ArgoCD, Prometheus/Grafana, and Loki) into your Kubernetes cluster.
```bash
cakd init
```

### Install specific components
To install only ArgoCD and Loki, explicitly specify their flags.
```bash
cakd init --argocd --logging
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)