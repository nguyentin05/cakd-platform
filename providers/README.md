# Declarative Provider Catalog

This directory contains the declarative provider specifications for the CAKD Platform.

## Architecture

The CAKD Platform uses a declarative provider catalog decoupled from the core Go binary.
Catalog manifests define provider metadata, dynamic configuration introspection endpoints, and interactive setup steps without hardcoding vendor-specific logic into the application core.

### Directory Structure

```
providers/
├── index.json                  # Catalog registry index listing available providers
├── README.md                   # Catalog documentation and authoring guide
└── vcs/                        # Version Control System providers
    ├── github.yaml             # GitHub VCS manifest with introspection
    └── local-git.yaml          # Local Git VCS manifest
```

## Manifest Schema

Each provider manifest YAML file contains:

- `id` (string): Unique identifier for the provider (e.g. `github`, `local-git`).
- `name` (string): Human-readable display name.
- `category` (string): Provider category (e.g. `vcs`, `ci`, `cd`, `notification`, `cluster`).
- `description` (string): Concise summary of the provider's capability.
- `version` (string): Manifest schema or provider version.
- `introspection` (optional object):
  - `fetch.url` (string): Endpoint for querying remote choices (supports `{{.token}}` interpolation).
  - `fetch.headers` (map): HTTP headers sent with the introspection request.
- `steps` (list of step objects):
  - `id` (string): Step key identifier.
  - `title` (string): Prompt label presented to user or caller.
  - `type` (string): Input type (e.g. `choice`, `string`).
  - `default` (optional string): Default value.
  - `extract` (optional string): JSONPath / `gjson` expression to extract dynamic choices from introspection response (e.g. `"$[*].login"`).
  - `options` (optional list): Static choices with `label` and `value`.

## Contributing a New Provider

1. Create a YAML manifest under the appropriate category directory (e.g., `providers/vcs/gitlab.yaml`).
2. Register the provider in `providers/index.json` under the `providers` list with its `id`, `name`, `category`, and `description`.
3. If the provider supports dynamic introspection, declare the `introspection` section with authenticated endpoints.
4. Test locally using `cakd-cli` or unit test suites.
