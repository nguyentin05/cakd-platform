# CLI Reference (`cakd`)

The `cakd` binary provides human-friendly terminal output by default and machine-readable JSON when `--json` is passed.

## Commands

| Command | Description |
| :--- | :--- |
| `cakd validate` | Validate `platform.yaml` syntax, enums, defaults, and dependency DAG |
| `cakd init` | Provision cloud host (optional) and bootstrap K3s, ArgoCD, Prometheus, Loki, and Agent |
| `cakd create` | Scaffold services, provision GitHub repository, push code, and register ArgoCD app |
| `cakd observe` | Collect Prometheus metrics + Loki logs and run Gemini AI root-cause diagnosis |
| `cakd auth` | Manage local provider credentials (`set`, `status`, `logout`) |
| `cakd schema` | Export Draft 2020-12 JSON Schema for `platform.yaml` |
| `cakd ui` | Launch the interactive TUI project creation wizard |
| `cakd mcp` | Start the Model Context Protocol (MCP) server (`--transport=stdio\|http`) |
| `cakd version` | Print build version |
