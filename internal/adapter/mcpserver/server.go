package mcpserver

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/platform/buildinfo"
	"github.com/nguyentin05/cakd-platform/internal/usecase/agenttools"
)

const (
	protocolVersion = "2025-03-26"
	jsonRPCVersion  = "2.0"
	contentTypeText = "text"
)

// Server serves Model Context Protocol (MCP) JSON-RPC 2.0 requests over Stdio or Streamable HTTP.
type Server struct {
	catalog *agenttools.Catalog
}

// NewServer creates an MCP Server backed by an agenttools.Catalog.
func NewServer(catalog *agenttools.Catalog) *Server {
	return &Server{catalog: catalog}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ServeStdio reads newline-delimited JSON-RPC 2.0 requests from in and writes responses to out.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	enc := json.NewEncoder(out)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		resp, shouldReply := s.handleRawJSON(ctx, []byte(line))
		if shouldReply {
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// NewHTTPHandler returns an http.Handler serving Streamable HTTP MCP on POST /mcp and GET /healthz.
func (s *Server) NewHTTPHandler(authToken string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if authToken != "" && !validBearerToken(r, authToken) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4*1024*1024)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		resp, shouldReply := s.handleRawJSON(r.Context(), body)
		if !shouldReply {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})
	return mux
}

func validBearerToken(r *http.Request, expected string) bool {
	hdr := r.Header.Get("Authorization")
	if !strings.HasPrefix(hdr, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(hdr, "Bearer ")
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

//nolint:gocyclo,goconst
func (s *Server) handleRawJSON(ctx context.Context, raw []byte) (rpcResponse, bool) {
	var req rpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			Error:   &rpcError{Code: -32700, Message: "parse error: " + err.Error()},
		}, true
	}

	if strings.HasPrefix(req.Method, "notifications/") {
		return rpcResponse{}, false
	}

	switch req.Method {
	case "initialize":
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": protocolVersion,
				"capabilities": map[string]any{
					"tools":     map[string]any{"listChanged": false},
					"resources": map[string]any{"listChanged": false},
				},
				"serverInfo": map[string]any{
					"name":    "cakd-mcp",
					"version": buildinfo.Version,
				},
			},
		}, true

	case "ping":
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Result:  map[string]any{},
		}, true

	case "tools/list":
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.catalog.ListTools(),
			},
		}, true

	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return rpcResponse{
				JSONRPC: jsonRPCVersion,
				ID:      req.ID,
				Error:   &rpcError{Code: -32602, Message: "invalid tool call params: " + err.Error()},
			}, true
		}
		text, err := s.catalog.CallTool(ctx, params.Name, params.Arguments)
		if err != nil {
			return rpcResponse{
				JSONRPC: jsonRPCVersion,
				ID:      req.ID,
				Result: map[string]any{
					"isError": true,
					"content": []map[string]any{
						{"type": contentTypeText, contentTypeText: err.Error()},
					},
				},
			}, true
		}
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Result: map[string]any{
				"isError": false,
				"content": []map[string]any{
					{"type": contentTypeText, contentTypeText: text},
				},
			},
		}, true

	case "resources/list":
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Result: map[string]any{
				"resources": s.catalog.ListResources(),
			},
		}, true

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return rpcResponse{
				JSONRPC: jsonRPCVersion,
				ID:      req.ID,
				Error:   &rpcError{Code: -32602, Message: "invalid resource read params: " + err.Error()},
			}, true
		}
		content, mime, err := s.catalog.ReadResource(ctx, params.URI)
		if err != nil {
			return rpcResponse{
				JSONRPC: jsonRPCVersion,
				ID:      req.ID,
				Error:   &rpcError{Code: -32602, Message: err.Error()},
			}, true
		}
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Result: map[string]any{
				"contents": []map[string]any{
					{
						"uri":           params.URI,
						"mimeType":      mime,
						contentTypeText: content,
					},
				},
			},
		}, true

	default:
		return rpcResponse{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
		}, true
	}
}
