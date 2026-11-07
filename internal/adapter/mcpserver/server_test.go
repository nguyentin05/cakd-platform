package mcpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/adapter/mcpserver"
	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func TestMCPServerStdioAndStreamableHTTP(t *testing.T) {
	c := composition.NewContainer()
	srv := mcpserver.NewServer(c.AgentTools)

	stdioReqs := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"resources/list","params":{}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"cakd_list_catalog","arguments":{}}}`,
	}, "\n") + "\n"

	var stdioOut bytes.Buffer
	if err := srv.ServeStdio(context.Background(), strings.NewReader(stdioReqs), &stdioOut); err != nil {
		t.Fatalf("ServeStdio failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdioOut.String()), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 JSON-RPC responses over stdio, got %d:\n%s", len(lines), stdioOut.String())
	}

	httpHandler := srv.NewHTTPHandler("secret-mcp-token")
	ts := httptest.NewServer(httpHandler)
	defer ts.Close()

	// Unauthorized request should receive 401
	unauthReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	resp, err := http.DefaultClient.Do(unauthReq)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing bearer token, got %d", resp.StatusCode)
	}

	// Authorized request should receive 200
	authReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	authReq.Header.Set("Authorization", "Bearer secret-mcp-token")
	authReq.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(authReq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for valid bearer token, got %d", resp2.StatusCode)
	}
	var rpcResp map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&rpcResp); err != nil {
		t.Fatal(err)
	}
	if rpcResp["result"] == nil {
		t.Fatalf("expected result in rpc response, got %+v", rpcResp)
	}
}
