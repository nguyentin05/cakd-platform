---
title: cakd init
description: Bootstrap the Kubernetes cluster with Platform Infrastructure (ArgoCD, Prometheus, Loki).
sidebar:
  order: 300
---

## Overview

The `cakd init` command bootstraps your Kubernetes cluster by installing essential platform infrastructure components: ArgoCD, Prometheus/Grafana, and Loki. Run this after you provision a cluster and before deploying `cakd` projects.

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

1. If no flags are passed, the command selects all components (ArgoCD, monitoring, logging).
2. For each selected component, it prepares helm commands and executes them sequentially.
3. When monitoring is selected, it writes a temporary Prometheus values file to configure Alertmanager to route to the CAKD Agent webhook.
4. The command streams output from helm and exits on failure.

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

- [Quickstart](/tutorials/quickstart/)
