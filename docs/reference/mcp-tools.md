# MCP Tools & Resources Reference

`cakd-cli mcp` and `cakd-mcp` expose the following Model Context Protocol (MCP) primitives:

## MCP Resources

| URI | MIME Type | Description |
| :--- | :--- | :--- |
| `cakd://schema/project` | `application/json` | Draft 2020-12 JSON Schema for `platform.yaml` |
| `cakd://catalog/providers` | `application/json` | Supported cluster providers, Java/Spring Boot versions, and defaults |

## MCP Tools

| Tool Name | Read-Only | Destructive | Description |
| :--- | :--- | :--- | :--- |
| `cakd_list_catalog` | Yes | No | List supported providers, versions, and defaults |
| `cakd_validate_spec` | Yes | No | Validate raw YAML or file path and return normalized spec |
| `cakd_generate_spec` | No | No | Generate canonical `cakd.io/v1alpha1` YAML from high-level parameters |
| `cakd_scaffold_only` | No | No | Render service skeletons, Dockerfiles, CI workflows, and Helm charts locally |
| `cakd_create_project` | No | Yes | Execute full project bootstrap (Terraform VCS + Git push + ArgoCD) |
| `cakd_init_cluster` | No | Yes | Initialize cluster and install Helm add-ons |
| `cakd_observe_workload` | Yes | No | Collect Prometheus metrics + Loki logs and run AI diagnosis |
