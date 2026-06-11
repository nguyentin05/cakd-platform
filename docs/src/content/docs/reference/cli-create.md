---
title: cakd create
description: The create command bootstraps a new cloud-native project based on a platform.yaml configuration.
sidebar:
  order: 100
---

## Overview

The `cakd create` command bootstraps a new cloud-native project. It reads `platform.yaml`, generates project files, provisions a GitHub repository via the Terraform Bridge, pushes the generated code, and registers an ArgoCD Application.

## Usage

```bash
cakd create [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
|`--file`|`-f`|`string`|`platform.yaml`|Path to the `platform.yaml` configuration file.|
|`--force`|`""`|`bool`|`false`|Overwrite the existing `project` directory if it already exists.|

## How It Works

1. Determine output directory (e.g., `./out/{project-name}`); fail if it exists unless `--force` is set.
2. Parse and validate `platform.yaml` using the three-phase pipeline (structure, defaults, logic).
3. Scaffold base project and render CAKD templates into the output directory.
4. Provision GitHub repository via the Terraform Bridge and retrieve outputs.
5. Initialize git, commit generated files, and push to the provisioned repository.
6. Register the ArgoCD application using the generated manifest.

## Examples

### Basic usage

```bash
cakd create -f platform.yaml
```

### Overwrite an existing project

```bash
cakd create -f platform.yaml --force
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)
