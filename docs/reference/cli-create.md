# `cakd create`

Scaffolds microservices, renders Dockerfiles, CI workflows, and Helm charts, provisions a GitHub repository via Terraform, pushes initial code, and registers the ArgoCD application.

## Usage

```bash
# Full pipeline
cakd create -f platform.yaml [--force] [--json]

# Scaffold locally only (no cloud/GitHub calls)
cakd create -f platform.yaml --scaffold-only [--force]

# Interactive TUI wizard
cakd create -i
```
