# ADR 003: Multi-Surface Driving Adapters (CLI, TUI & MCP)

- **Status**: Accepted
- **Date**: 2026-11-07

## Context

Modern platform engineering serves both human operators (who benefit from an interactive wizard or CLI) and AI coding agents (which interact via the Model Context Protocol).

## Decision

Expose the same application use cases (`internal/usecase/*`) through three primary driving surfaces:

1. **CLI (`internal/adapter/cli`)** with `--json` output mode.
2. **Interactive TUI Wizard (`internal/adapter/tui`)** via `cakd-cli ui` and `cakd-cli create -i`.
3. **MCP Server (`internal/adapter/mcpserver`)** via `cakd-cli mcp` and standalone `cakd-mcp` binary supporting both Stdio and Streamable HTTP transports.
