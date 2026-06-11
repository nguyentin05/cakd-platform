---
title: cakd create
description: The create command bootstraps a new cloud-native project based on a platform.yaml configuration.
sidebar:
  order: 100
---

## Overview

The `cakd create` command is used to `bootstrap` a new cloud-native project. It takes a `platform.yaml` configuration file as input, generates the project structure, provisions a GitHub repository, pushes the initial code, and registers the application with ArgoCD. Use this command to quickly set up a new service ready for GitOps deployment.

## Usage

```bash
cakd create [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
|`--file`|`-f`|`string`|`""`|Path to the `platform.yaml` configuration file.|
|`--force`|`""`|`bool`|`false`|Overwrite the existing `project` directory if it already exists.|

## How It Works

1.  The command first determines the output directory, typically `./out/{project-name}`.
2.  If the output directory already exists and the `--force` flag is not used, the command returns an error. If `--force` is specified, the existing directory is removed.
3.  The output directory is created.
4.  **`platform.yaml` Parsing and Validation:**
    *   The `platform.yaml` file specified by the `--file` flag is read from the filesystem.
    *   The YAML content is unmarshaled into an internal configuration structure.
    *   The configuration then undergoes a strict, 3-phase processing pipeline:
        1.  **Structure Validation:** Scans the raw YAML to ensure all required fields and enum values are present, throwing errors before any mutation occurs.
        2.  **Defaults Injection:** Recursively applies implicit defaults (Convention over Configuration) based on tags and internal registry values for missing fields.
        3.  **Logic/Dependency Validation:** Evaluates cross-field dependencies and business rules (e.g., CI requires CD) now that the configuration tree is fully hydrated with defaults.
5.  **Project Generation (Step 1/4):**
    *   If the `platform.yaml` specifies `java-spring-boot` as the language, a base project is downloaded from `start.spring.io`.
    *   The `template engine` then applies `cakd`-specific templates, including Dockerfile, Helm charts, CI/CD pipelines, and ArgoCD manifests, into the output directory.
6.  **GitHub Repository Creation (Step 2/4):**
    *   The `Terraform Bridge` is invoked to provision a new GitHub repository based on the project configuration.
7.  **Code Push (Step 3/4):**
    *   The generated `project` code is initialized as a Git repository and pushed to the newly created GitHub repository.
    *   If the Git push fails, a rollback mechanism attempts to destroy the Terraform-provisioned GitHub repository.
8.  **ArgoCD Registration (Step 4/4):**
    *   The ArgoCD application manifest (located at `deploy/application.yaml` within the `project`) is registered with an ArgoCD instance.
9.  Finally, a success message is displayed, including the GitHub repository URL and the local `project` path.

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