---
title: cakd observe
description: Use AI to observe, diagnose, and troubleshoot a project.
sidebar:
  order: 100
---

## Overview

The `cakd observe` command uses AI to analyze metrics and logs for a `project` namespace and return a concise diagnosis and remediation steps.

## Usage

```bash
cakd observe <project-name> [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|------|------|--------|------------|

## How It Works

1. Accept a single argument: the `project` (Kubernetes namespace).
2. Read `GEMINI_API_KEY` (via config helper). Exit with error if unset.
3. Initialize Prometheus, Loki, and Gemini clients.
4. Fetch pod/container restart metrics from Prometheus and recent logs from Loki.
5. Build a detailed prompt containing metrics and logs and send it to Gemini.
6. Print the AI-generated diagnosis to stdout.

## Examples

```bash
export GEMINI_API_KEY="your_api_key"
cakd observe my-app
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)
