---
title: cakd observe
description: Use AI to observe, diagnose, and troubleshoot a project.
sidebar:
  order: 100
---

## Overview

The `cakd observe` command leverages artificial intelligence to provide deep insights into the health and performance of a running `project` on Kubernetes. Use this command to automatically diagnose issues, identify root causes, and receive actionable solutions for your applications. It produces an AI-generated diagnosis and solution, presented in Vietnamese, based on real-time metrics and logs from your cluster.

## Usage

```bash
cakd observe <project-name> [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|------|------|--------|------------|

## How It Works

1.  The command takes a `project` name as its sole argument, which corresponds to a Kubernetes namespace.
2.  It checks for the `GEMINI_API_KEY` environment variable. If this variable is not set, the command exits with an error, instructing you to set it.
3.  It initializes internal clients for fetching metrics from Prometheus, logs from Loki, and for interacting with the Google Gemini AI.
4.  The command then fetches metrics related to pod container restarts from Prometheus for the specified `project` namespace by executing `kubectl get --raw` against the Prometheus service.
5.  Concurrently, it fetches recent logs from Loki for the `project` namespace, also by executing `kubectl get --raw` against the Loki service.
6.  It constructs a detailed prompt for the AI, incorporating the fetched metrics and logs, and instructing the AI to act as an expert DevOps engineer and to provide the diagnosis in Vietnamese.
7.  This prompt is sent to the Google Gemini API (`gemini-flash-latest`) via an HTTP POST request for analysis.
8.  The AI's diagnosis, which includes identifying crashing pods, determining root causes, and providing actionable solutions, is then printed directly to your console.

## Examples

### Basic usage
```bash
cakd observe my-app-project
```

### Setting the API key
```bash
export GEMINI_API_KEY="your_api_key"
cakd observe my-app-project
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)