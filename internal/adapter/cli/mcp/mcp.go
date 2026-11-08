// Package mcp implements the `cakd-cli mcp` subcommand.
package mcp

import (
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/adapter/mcpserver"
)

// NewCommand creates the `cakd-cli mcp` subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		transport string
		addr      string
		authToken string
	)
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the Model Context Protocol (MCP) server over stdio or Streamable HTTP",
		RunE: func(cmd *cobra.Command, _ []string) error {
			srv := mcpserver.NewServer(f.C.AgentTools)
			if transport == "http" {
				if authToken == "" {
					authToken = os.Getenv("CAKD_MCP_AUTH_TOKEN")
				}
				httpSrv := &http.Server{
					Addr:              addr,
					Handler:           srv.NewHTTPHandler(authToken),
					ReadHeaderTimeout: 10 * time.Second,
				}
				return httpSrv.ListenAndServe()
			}
			return srv.ServeStdio(cmd.Context(), f.In, f.Out)
		},
	}
	cmd.Flags().StringVar(&transport, "transport", "stdio", "MCP transport: stdio or http")
	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:8090", "Listen address when --transport=http")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Bearer token for HTTP MCP transport (or CAKD_MCP_AUTH_TOKEN)")
	return cmd
}
