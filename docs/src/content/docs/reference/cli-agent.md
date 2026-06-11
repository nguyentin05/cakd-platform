---
title: cakd-agent
description: The background diagnostics agent that listens for Alertmanager webhooks.
sidebar:
  order: 250
---

## Overview

The `cakd-agent` daemon is a real-time diagnostics runner. It listens for Alertmanager webhooks, collects system logs and metrics from Loki and Prometheus, and invokes Gemini to diagnose failures before routing root-cause alerts to Discord.

## Usage

```bash
cakd-agent
```

## Environment Variables

| Env Var | Type | Default | Description |
|---------|------|---------|-------------|
|`DISCORD_WEBHOOK_URL`|`string`|`""`|The webhook URL to post diagnostics reports to|
|`PORT`|`string`|`"8080"`|The port where the agent daemon listens for Alertmanager webhooks|

## How It Works

1. `cakd-agent` starts an HTTP server listening on the configured `PORT`.
2. It registers the `/api/v1/alerts` endpoint to receive webhooks forwarded by Prometheus Alertmanager.
3. When an alert arrives, the agent queries metrics (via Prometheus) and log streams (via Loki) to construct failure context.
4. It submits the log traces and metrics context to the Gemini LLM service to compile a diagnostic analysis of the failure.
5. Finally, it formats a rich embed card containing the diagnosis, logs summary, and metric indicators, and delivers it to Discord.

## Examples

### Basic usage

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
cakd-agent
```

## Related

- [Quickstart tutorial](/tutorials/quickstart/)
- [How to set up Discord alerts](/how-to-guides/setup-discord-alerts/)
