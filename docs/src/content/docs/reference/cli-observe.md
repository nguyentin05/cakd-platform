---
title: `cakd` observe
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

## Environment Variables

| Env Var | Short | Type | Default | Description |
|---------|-------|------|---------|-------------|
|`GEMINI_API_KEY`| |`string`|`""`|Your Google Gemini API key. Required for AI diagnosis.|
|`GEMINI_MODEL`| |`string`|`"gemini-flash-latest"`|The Gemini model to use for AI diagnosis.|

## How It Works

1.  Accepts a single argument: the `project` (Kubernetes namespace).
2.  Reads the `GEMINI_API_KEY` environment variable. If unset, the command exits with an error.
3.  Initializes clients for Prometheus (metrics), Loki (logs), and Gemini (AI). The Gemini client uses the `GEMINI_MODEL` environment variable, defaulting to `gemini-flash-latest` if not specified.
4.  Fetches container restart metrics for the specified `project` from Prometheus. If fetching fails, a warning is printed, and "No metrics available" is used in the prompt.
5.  Fetches the 50 most recent logs for the specified `project` from Loki. If fetching fails, a warning is printed, and "No logs available" is used in the prompt.
6.  Constructs a detailed diagnostic prompt for the AI, including the fetched metrics and logs, and instructions to act as an expert DevOps engineer. The prompt explicitly requests the AI to "Answer in Vietnamese."
7.  Sends the prompt to the Gemini AI model for analysis.
8.  Prints the AI-generated diagnosis, including identification of crashing pods, root cause analysis, and actionable solutions, to standard output.

## Examples

### Basic usage
```bash
export GEMINI_API_KEY="your_api_key"
cakd observe my-app
```

### Specifying a Gemini model
```bash
export GEMINI_API_KEY="your_api_key"
export GEMINI_MODEL="gemini-pro"
cakd observe my-app
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)