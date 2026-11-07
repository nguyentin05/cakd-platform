# MCP Agent Integration

CAKD Platform ships with a native Model Context Protocol (MCP) server (`cakd mcp` or `cakd-mcp`) supporting both **Stdio** and **Streamable HTTP** transports.

## Stdio Configuration (Claude Desktop / Cursor / Antigravity)

Add the following entry to your MCP client configuration:

```json
{
  "mcpServers": {
    "cakd": {
      "command": "/path/to/bin/cakd-mcp",
      "args": ["--transport=stdio"]
    }
  }
}
```

## Streamable HTTP Mode

```bash
export CAKD_MCP_AUTH_TOKEN="my-secret-token"
cakd mcp --transport=http --addr=127.0.0.1:8090
```

Clients can then send JSON-RPC 2.0 `POST /mcp` requests with `Authorization: Bearer my-secret-token`.
