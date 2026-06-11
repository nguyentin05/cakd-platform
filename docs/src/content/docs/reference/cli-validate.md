---
title: cakd validate
description: Validates a platform.yaml file against the CAKD Platform schema.
sidebar:
  order: 200
---

## Overview

The `cakd validate` command checks a `platform.yaml` file for structural and semantic correctness. Use it to verify configuration before running `cakd create`.

## Usage

```bash
cakd validate [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
|`--file`|`-f`|`string`|`platform.yaml`|Path to `platform.yaml`|

## How It Works

1. Parse the `platform.yaml` file using the config parser.
2. Perform structure and logic validation. Exit non-zero on validation errors.
3. On success, print project metadata (name, owner, language).

## Examples

```bash
cakd validate -f platform.yaml
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
