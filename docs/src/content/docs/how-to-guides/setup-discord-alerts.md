---
title: How to Setup Discord Alerts
description: This guide explains how to configure Discord as a notification provider in platform.yaml to receive automated alerts and system diagnostics.
---

## Overview

The CAKD Platform supports pluggable notification providers. This guide explains how to configure and deploy Discord alerting so that your project pipeline and the background diagnostics agent (`cakd-agent`) can deliver real-time notifications to your Discord channel.

## Provisioning Discord alerting (Notify provider)

When `platform.yaml` configures a notification provider, the pipeline runs the `NotifyStep` to provision notification channels.

1. The pipeline invokes the notification provider factory with the configured provider name.
2. For Discord, the provider implementation uses the Discord REST API to create a channel and webhook (`internal/provider/notify/discord`).
3. The webhook URL is returned to the pipeline, printed to stdout, and saved to local config for routing.

## Example: Configure Discord as the notification provider

1. Ensure you have a Discord bot token and guild ID. Keep them private.
2. Add provider selection in your `platform.yaml` (example snippet):

```yaml
providers:
  notification: discord
```

3. Run the create pipeline:

```bash
cakd create -f platform.yaml
```

4. The pipeline will attempt to create a Discord channel and webhook for the project. If it succeeds, the webhook URL is printed and stored locally.

## When provisioning fails

The pipeline treats notify provisioning as optional. If the provider returns an error, the pipeline prints a warning and continues. Check logs for API errors and validate bot permissions (creating channels, managing webhooks).

## Related

- [`cakd create` reference](/reference/cli-create/)
- [Quickstart](/tutorials/quickstart/)
