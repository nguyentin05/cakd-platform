---
title: CLI Authentication
description: How to authenticate providers (GitHub, Discord, Gemini) using the cakd CLI, environment variables, and the local credentials file.
sidebar:
  parent: Reference
  order: 20
---

## Overview

The `cakd` CLI provides a small credentials manager that securely stores and retrieves tokens and keys for the platform's providers (GitHub, Discord, Gemini). Use the `auth` command family to interactively save credentials, inspect the authentication status, and remove stored credentials.

Supported providers:
- github — GitHub Personal Access Token (PAT)
- discord — Discord Bot Token + Guild ID
- gemini — Google Gemini API key

The CLI integrates with environment variables and a local credentials file (~/.cakd/credentials.yaml) and follows a deterministic resolution order described below.

## Commands

### auth login [provider]

Interactively prompts for the provider's credentials and optionally saves them to the local credentials store.

Examples:

- Log in to GitHub:
```bash
cakd auth login github
# prompts: "Enter GitHub Personal Access Token (PAT):" (input hidden)
```

- Log in to Discord (bot token + guild id):
```bash
cakd auth login discord
# prompts: "Enter Discord Bot Token:" (input hidden)
# then: "Enter Discord Guild ID (Server ID):"
```

- Log in to Gemini:
```bash
cakd auth login gemini
# prompts: "Enter Gemini API Key:" (input hidden)
```

Notes:
- Masked input (no echo) is used for tokens/keys.
- After entering credentials you will be offered to persist them to `~/.cakd/credentials.yaml` (interactive save).

### auth status

Show current authentication status for all providers. Tokens are masked when displayed.

Example output:
```
Credentials file path: /home/alice/.cakd/credentials.yaml

Provider Authentication Status:
  GitHub:  ✅ Authenticated (ends in ...3f2a)
  Discord: ✅ Authenticated (Guild ID: 123456789012345678)
  Gemini:  ❌ Not Authenticated
```

### auth logout [provider]

Remove stored credentials for the given provider.

Example:
```bash
cakd auth logout github
# Removes saved GitHub token from ~/.cakd/credentials.yaml
```

## Resolution order (how tokens are found)

When a `cakd` command needs provider credentials, the code resolves values in the following order for each provider:

GitHub token:
1. CAKD_GITHUB_TOKEN environment variable
2. GITHUB_TOKEN environment variable
3. Local credentials file: `~/.cakd/credentials.yaml` (field: `github.token`)
4. Interactive prompt (only if stdin is a terminal; you will be offered to save)

Discord token and guild id:
1. CAKD_DISCORD_BOT_TOKEN or DISCORD_BOT_TOKEN
2. CAKD_DISCORD_GUILD_ID or DISCORD_GUILD_ID
3. Local credentials file: `discord.token` and `discord.guild_id`
4. Interactive prompt (if interactive)

Gemini API key:
1. CAKD_GEMINI_API_KEY or GEMINI_API_KEY
2. Local credentials file: `gemini.api_key`
3. Interactive prompt (if interactive)

If a credential is not resolved and the process is not interactive (e.g., CI), the operation will fail and instruct you to provide credentials via environment variables or the local credentials file.

## Local credentials file

Credentials are optionally stored in YAML at:

`~/.cakd/credentials.yaml`

The file is created with directory permissions 0700 and file permissions 0600 (user-only access). Example structure:

```yaml
github:
  token: "ghp_0123456789example"
discord:
  token: "Bot abcdef...example"
  guild_id: "123456789012345678"
gemini:
  api_key: "AIza...example"
```

Important:
- Do not commit this file to git.
- The CLI will not save credentials unless you confirm the interactive prompt.

## Non-interactive usage / CI

For automation and CI pipelines, prefer environment variables:

- GitHub: set CAKD_GITHUB_TOKEN or GITHUB_TOKEN
- Discord: set CAKD_DISCORD_BOT_TOKEN and CAKD_DISCORD_GUILD_ID
- Gemini: set CAKD_GEMINI_API_KEY or GEMINI_API_KEY

Environment variables take precedence over values stored in `~/.cakd/credentials.yaml`.

## Security recommendations

- Use environment variables or a secret manager (CI) in automation rather than storing secrets in plain files.
- If storing credentials locally during development, rely on the CLI to write to `~/.cakd/credentials.yaml` which is created with restrictive permissions (0600).
- Rotate tokens regularly and only grant the minimal scopes required by the platform (for GitHub, prefer least-privilege PAT scopes).
- Never paste secrets into public issues or commit them into source control.

## Troubleshooting

- "stdin is not a terminal, cannot prompt securely": This means `cakd` attempted to prompt but stdin is not interactive. Use environment variables or create `~/.cakd/credentials.yaml` manually.
- Token present but commands still fail: verify the token has the required scopes (e.g., repo) and that there are no leading/trailing whitespace characters.
- To inspect the stored file path and masked status:
```bash
cakd auth status
```

## Example: programmatic (CI) pattern

In a GitHub Actions job you might provide credentials via secrets:
```yaml
env:
  CAKD_GITHUB_TOKEN: ${{ secrets.CAKD_GITHUB_TOKEN }}
  CAKD_GEMINI_API_KEY: ${{ secrets.CAKD_GEMINI_API_KEY }}
steps:
  - uses: actions/checkout@v4
  - name: Run cakd create
    run: cakd create -f platform.yaml
```

This ensures the CLI picks up credentials without prompting.
