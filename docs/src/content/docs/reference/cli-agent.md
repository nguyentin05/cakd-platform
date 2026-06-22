---
title: cakd-agent
description: The background diagnostics agent that listens for Alertmanager webhooks.
sidebar:
  order: 250
---

## Overview

The `` `cakd-agent` `` daemon is a real-time diagnostics runner. It listens for Alertmanager webhooks, processes incoming alerts, invokes Gemini to diagnose failures based on alert descriptions, and routes both raw alerts and AI-generated root-cause analyses to Discord.

## Usage

```bash
cakd-agent
```

## Environment Variables

| Env Var | Type | Default | Description |
|---------|------|---------|-------------|
|`DISCORD_WEBHOOK_URL`|`string`|`""`|The webhook URL to post diagnostics reports to|
|`PORT`|`string`|`"8080"`|The port where the agent daemon listens for Alertmanager webhooks|
|`GEMINI_API_KEY`|`string`|`""`|The API key for authenticating with the Gemini LLM service|
|`GEMINI_MODEL`|`string`|`"gemini-flash-latest"`|The Gemini LLM model to use for AI analysis|
|`CAKD_AGENT_SECRET`|`string`|`""`|A shared secret for authenticating incoming Alertmanager webhooks|

## How It Works

1.  `` `cakd-agent` `` starts an HTTP server listening on the configured `PORT`.
2.  It registers the `/api/v1/alerts` endpoint to receive webhooks forwarded by Prometheus Alertmanager.
3.  Upon receiving a POST request, it validates the `Content-Type` header and limits the request body size to 1MB. If `CAKD_AGENT_SECRET` is configured, it validates the shared secret from the `Authorization` header (expecting a `Bearer` token) or `secret` query parameter.
4.  The agent decodes the incoming Alertmanager JSON payload.
5.  It immediately sends an HTTP `200 OK` response and then processes the alerts asynchronously.
6.  It groups the received alerts by their target webhook URL, which can be the `DISCORD_WEBHOOK_URL` or a namespace-specific URL loaded from configuration.
7.  For each group, it formats the raw alert details (status, name, severity, description, defaulting to "No description provided" if empty) and dispatches them via the configured notifier (e.g., Discord).
8.  If there are "firing" alerts and a Gemini client is initialized (i.e., `GEMINI_API_KEY` is set), the agent asynchronously sends the collected alert descriptions to the Gemini LLM service for AI analysis.
9.  The Gemini LLM generates a concise diagnosis and troubleshooting steps based on the provided alert context.
10. The AI-generated diagnosis, titled "🤖 CAKD AI Diagnosis", is then formatted as an informational alert and sent back to the relevant target webhook URL via the notifier.

## Examples

### Basic usage

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
cakd-agent
```

### With AI analysis and webhook authentication

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export GEMINI_API_KEY="your-gemini-api-key"
export CAKD_AGENT_SECRET="your-shared-secret"
cakd-agent
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)
- [How to set up Discord alerts](/how-to-guides/setup-discord-alerts/)