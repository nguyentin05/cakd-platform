package mcpserver

import "github.com/nguyentin05/cakd-platform/internal/usecase/agenttools"

// ToolAnnotations defines MCP protocol annotations for tool execution safety hints.
type ToolAnnotations = agenttools.ToolHints

// ToolDefinition defines an MCP tool definition serialized over JSON-RPC 2.0.
type ToolDefinition = agenttools.ToolDescriptor

// ResourceDefinition defines an MCP resource definition serialized over JSON-RPC 2.0.
type ResourceDefinition = agenttools.ResourceDescriptor
