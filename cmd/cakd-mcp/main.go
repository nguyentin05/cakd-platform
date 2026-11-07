package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nguyentin05/cakd-platform/internal/adapter/mcpserver"
	"github.com/nguyentin05/cakd-platform/internal/composition"
)

func main() {
	var (
		transport string
		addr      string
		authToken string
	)
	flag.StringVar(&transport, "transport", "stdio", "MCP transport: stdio or http")
	flag.StringVar(&addr, "addr", "127.0.0.1:8090", "Listen address when --transport=http")
	flag.StringVar(&authToken, "auth-token", "", "Bearer token for HTTP MCP transport (or CAKD_MCP_AUTH_TOKEN)")
	flag.Parse()

	container := composition.NewContainer()
	srv := mcpserver.NewServer(container.AgentTools)

	if transport == "http" {
		if authToken == "" {
			authToken = os.Getenv("CAKD_MCP_AUTH_TOKEN")
		}
		httpSrv := &http.Server{
			Addr:              addr,
			Handler:           srv.NewHTTPHandler(authToken),
			ReadHeaderTimeout: 10 * time.Second,
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP HTTP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := srv.ServeStdio(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "MCP Stdio server error: %v\n", err)
		os.Exit(1)
	}
}
