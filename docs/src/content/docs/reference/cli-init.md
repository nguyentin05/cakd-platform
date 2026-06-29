---
title: cakd init
description: Bootstrap the Kubernetes cluster with Platform Infrastructure (ArgoCD, Prometheus, Loki).
sidebar:
  order: 300
---

## Overview

The `cakd init` command bootstraps your Kubernetes cluster by installing essential platform infrastructure components: ArgoCD, Prometheus/Grafana, Loki, and the `cakd-agent`. Run this after you provision a cluster and before deploying `cakd` projects.

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

1.  The command first checks if any component flags (`--argocd`, `--monitoring`, `--logging`) are provided. If none are specified, it defaults to installing all three components.
2.  For each selected component, `cakd init` performs the following steps:
    *   **ArgoCD**: It adds the `argo` Helm repository, updates it, and then installs or upgrades ArgoCD into the `argocd` namespace.
    *   **Monitoring (Prometheus Stack)**:
        *   It creates a temporary Prometheus values file, embedding configuration for Alertmanager to route to the `cakd-agent` webhook.
        *   It adds the `prometheus-community` Helm repository, updates it, and then installs or upgrades the `kube-prometheus-stack` into the `monitoring` namespace, using the temporary values file.
    *   **Deploy CAKD Agent**: If monitoring is enabled, the `cakd-agent` is deployed to the cluster.
    *   **Logging (Loki)**: It adds the `grafana` Helm repository, updates it, and then installs or upgrades `loki-stack` into the `monitoring` namespace.
3.  All Helm commands and agent deployment steps are executed sequentially, and their output is streamed to the console.
4.  The command exits with an error if any step fails.

## Examples

### Install all default components

```bash
cakd init
```

### Install only ArgoCD and Loki

```bash
cakd init --argocd --logging
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)