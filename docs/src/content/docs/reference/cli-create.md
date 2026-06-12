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
|`--file`|`-f`|`string`|`platform.yaml`|Path to `platform.yaml`.|
|`--force`|`""`|`bool`|`false`|Force overwrite existing `project` directory.|

## How It Works

1.  The `platform.yaml` file is read from the specified path.
2.  The configuration is parsed and validated through a three-phase pipeline: structure validation, defaults injection, and logic validation.
3.  A project context is prepared, which includes determining the output directory (e.g., `./out/{project-name}`). If the directory already exists and `--force` is not set, the command fails.
4.  The following sequential pipeline steps are executed:
    *   **Scaffold**: Scaffolds the base `project` and renders CAKD templates into the output directory.
    *   **Infra**: Provisions infrastructure (e.g., a GitHub repository via the `Terraform Bridge`) and retrieves outputs.
    *   **Version Control**: Initializes git, commits generated files, and pushes them to the provisioned repository.
    *   **Deploy**: Registers the ArgoCD application using the generated manifest.
    *   **Notify**: Performs any final notification actions.
5.  Upon successful completion, a summary is printed, detailing the newly created `project`'s git repository location, local code path, and deployment mode.

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

-   [`platform.yaml` reference](/reference/platform-yaml/)
-   [Quickstart tutorial](/tutorials/quickstart/)