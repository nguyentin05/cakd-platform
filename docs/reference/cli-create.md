# `cakd-cli create`

Scaffolds microservices, renders Dockerfiles, CI workflows, and Helm charts, provisions a GitHub repository via Terraform, pushes initial code, and registers the ArgoCD application.

## Usage

```bash
# Full pipeline
cakd-cli create -f platform.yaml [--force] [--json]

# Scaffold locally only (no cloud/GitHub calls)
cakd-cli create -f platform.yaml --scaffold-only [--force]

# Interactive TUI wizard
cakd-cli create -i
```
