---
title: cakd validate
description: Validates a platform.yaml file against the CAKD Platform schema.
sidebar:
  order: 200
---

## Overview

The `cakd validate` command checks a `platform.yaml` file for structural and semantic correctness against the CAKD Platform schema. Use this command to verify your configuration before running `cakd create` or after making changes to your project's `platform.yaml`. It ensures your project definition is valid and ready for bootstrapping.

## Usage

```bash
cakd validate [flags]
```

## Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
|`--file`|`-f`|`string`|`"platform.yaml"`|Path to `platform.yaml`|

## How It Works

1.  `cakd validate` attempts to parse the `platform.yaml` file specified by the `--file` flag (or `platform.yaml` in the current directory if no flag is provided).
2.  It unmarshals the YAML content into the internal `PlatformConfig` structure, validating the data against the expected schema.
3.  If any structural or semantic errors are found during parsing, `cakd` prints a detailed error message to `stderr` and exits with a non-zero status code.
4.  If the `platform.yaml` file is valid, `cakd` prints a success message to `stdout`, including the project's name, owner, and the language of the first service extracted from the configuration.

## Examples

### Basic usage
```bash
cakd validate -f platform.yaml
```

### Validate a custom configuration file
```bash
cakd validate --file config/my-project.yaml
```

## Related

- [`platform.yaml` reference](/reference/platform-yaml/)
- [Quickstart tutorial](/tutorials/quickstart/)