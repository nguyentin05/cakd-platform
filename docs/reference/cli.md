# CLI Reference (`cakd-cli`)

The `cakd-cli` binary provides human-friendly terminal output by default and machine-readable JSON when `--json` is passed.

## Commands

| Command | Description |
| :--- | :--- |
| `cakd-cli validate` | Validate `platform.yaml` syntax, enums, defaults, and dependency DAG |
| `cakd-cli init` | Provision cloud host (optional) and bootstrap K3s, ArgoCD, Prometheus, Loki, and Agent |
| `cakd-cli create` | Scaffold services, provision GitHub repository, push code, and register ArgoCD app |
| `cakd-cli observe` | Collect Prometheus metrics + Loki logs and run Gemini AI root-cause diagnosis |
| `cakd-cli auth` | Manage local provider credentials (`set`, `status`, `logout`) |
| `cakd-cli schema` | Export Draft 2020-12 JSON Schema for `platform.yaml` |
| `cakd-cli ui` | Launch the interactive TUI project creation wizard |
| `cakd-cli mcp` | Start the Model Context Protocol (MCP) server (`--transport=stdio\|http`) |
| `cakd-cli version` | Print build version |
