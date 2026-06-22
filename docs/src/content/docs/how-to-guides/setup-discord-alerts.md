---
title: How to Setup Discord Alerts
description: This guide explains how to configure Discord as a notification provider in platform.yaml to receive automated alerts and system diagnostics.
---

## Overview

The CAKD Platform supports pluggable notification providers. This guide explains how to configure and deploy Discord alerting so that your `project` pipeline and the background diagnostics agent (`cakd-agent`) can deliver real-time notifications to your Discord channel.

## Prerequisites

- A Discord bot token.
- A Discord guild (server) ID.
- The Discord bot must have permissions to `Manage Channels` (to create text channels) and `Manage Webhooks` (to create webhooks) in the target guild.
- A `project` created with `cakd create`.

## Provisioning Discord alerting (Notify provider)

When `platform.yaml` configures a notification provider, the pipeline runs the `NotifyStep` to provision notification channels.

1.  The pipeline invokes the notification provider factory with the configured provider name.
2.  For Discord, the provider implementation uses the Discord REST API to create a channel and webhook (`internal/provider/notify/discord`).
3.  The webhook URL is returned to the pipeline, printed to stdout, and saved to local config for routing.

## Steps

### 1. Obtain Discord Bot Token and Guild ID

You need a Discord bot token and the ID of the Discord server (guild) where you want alerts to be posted.

:::caution Important
The Discord bot must have the following permissions in the target guild:
- `Manage Channels` (to create the `alerts-<project-name>` text channel)
- `Manage Webhooks` (to create the `CAKD Alerting - <project-name>` webhook)
:::

### 2. Configure `platform.yaml`

Specify `discord` as your notification provider in your `platform.yaml` file.

```yaml
providers:
  notification: discord
```

### 3. Run the `cakd create` command

Execute `cakd create` to provision the Discord channel and webhook for your `project`.

```bash
cakd create -f platform.yaml
```

## When provisioning fails

The pipeline treats notify provisioning as optional. If the provider returns an error, the pipeline prints a warning and continues. Check logs for API errors and validate bot permissions (creating channels, managing webhooks).

## Verify it works

Run the `cakd create` command again if you need to see the output, or check your Discord server.

```bash
cakd create -f platform.yaml
```

:::tip Expected result
You should see output similar to this during the `Provisioning Notification Channels` step:

```text
   Creating Discord channel for my-project...
   Creating Discord webhook...
   Webhook created: https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz
   Webhook routing rule saved for namespace: my-project
```
A new text channel named `alerts-my-project` will be created in your Discord guild, and a webhook named `CAKD Alerting - my-project` will be associated with it.
:::

## Troubleshooting

| Problem | Cause | Fix |
|---------|-------|-----|
| `Warning: Failed to provision notification channel: discord API error: status 403, body: {"message": "Missing Permissions", ...}` | The Discord bot token lacks necessary permissions (e.g., `Manage Channels`, `Manage Webhooks`) in the target guild. | Grant the bot the required permissions in the Discord Developer Portal. |
| `Warning: Failed to provision notification channel: failed to create channel: failed to make request: Post "https://discord.com/api/v10/guilds/.../channels": dial tcp ...` | Network connectivity issues preventing the CAKD Platform from reaching the Discord API. | Check your network connection, firewall rules, or Discord API status page. |
| `Warning: Failed to save webhook to local config: open /path/to/cakd/config/my-project.json: permission denied` | The user running `cakd` does not have write permissions to the local CAKD configuration directory. | Ensure the user has write permissions to the CAKD config directory (e.g., `~/.config/cakd/`). |

## Related

- [`cakd create` reference](/reference/cli-create/)
- [Quickstart](/tutorials/quickstart/)