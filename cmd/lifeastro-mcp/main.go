// Command lifeastro-mcp is the Model Context Protocol (MCP) server for
// LifeAstroAPI. It exposes a curated subset of /v1/* endpoints as MCP
// tools so AI assistants — Claude Desktop, Cursor, Continue.dev, any
// MCP-aware client — can call them directly without writing code.
//
// Architecture:
//
//	┌─────────────────────────┐
//	│  USER LAPTOP            │
//	│  Claude Desktop / Cursor│
//	│        ↓ stdio          │
//	│  lifeastro-mcp (this)   │
//	│        ↓ HTTPS          │
//	└────────│────────────────┘
//	         ↓ public internet
//	  api.lifeastroapi.com
//
// This binary runs locally as a subprocess of the AI client (stdio
// transport). It does NOT add load to our API server beyond the normal
// HTTP requests it forwards — every tool call becomes one authenticated
// HTTPS GET to api.lifeastroapi.com, identical to a curl call.
//
// Authentication is via the LIFEASTRO_API_KEY environment variable,
// which the AI client injects into the subprocess at spawn (configured
// in claude_desktop_config.json or .cursor/mcp.json). The API key is
// the standard dv_live_<hex> key issued via the existing admin API.
//
// Usage:
//
//	# In Claude Desktop config:
//	{
//	  "mcpServers": {
//	    "lifeastro": {
//	      "command": "/usr/local/bin/lifeastro-mcp",
//	      "env": { "LIFEASTRO_API_KEY": "dv_live_..." }
//	    }
//	  }
//	}
//
// Build:
//
//	go build -o lifeastro-mcp ./cmd/lifeastro-mcp/
//	# Or for a release:
//	go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD)" \
//	  -o lifeastro-mcp ./cmd/lifeastro-mcp/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is set at build time via -ldflags. Defaults to "dev" for
// local builds. Surfaced in the MCP Implementation handshake so AI
// clients can log which version is responding.
var version = "dev"

// commit is the git commit hash, also injected at build time. "none"
// for local builds.
var commit = "none"

// Default API base URL. Override via LIFEASTRO_API_URL env (useful for
// local development against http://localhost:8080).
const defaultAPIBaseURL = "https://api.lifeastroapi.com"

func main() {
	// Read auth from env. Fail fast — there's no recovery path if the
	// AI client hasn't injected the key, and silent fallback would hide
	// the configuration mistake from the user.
	apiKey := os.Getenv("LIFEASTRO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "lifeastro-mcp: LIFEASTRO_API_KEY env var is required.")
		fmt.Fprintln(os.Stderr, "Add it to your MCP client config:")
		fmt.Fprintln(os.Stderr, `  "env": { "LIFEASTRO_API_KEY": "dv_live_..." }`)
		fmt.Fprintln(os.Stderr, "Get a key at https://lifeastroapi.com/dashboard/keys")
		os.Exit(1)
	}

	apiBaseURL := os.Getenv("LIFEASTRO_API_URL")
	if apiBaseURL == "" {
		apiBaseURL = defaultAPIBaseURL
	}

	// One shared HTTP client for all upstream calls. Pooled connections
	// matter — a chatty AI assistant may fire 5+ tool calls per turn.
	client := newAPIClient(apiBaseURL, apiKey, version)

	// MCP server identity surfaces in the AI client's tool list and
	// debug logs. Keep the name short and the description specific.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "lifeastro",
		Version: version,
		Title:   "LifeAstroAPI — Vedic & Western Astrology",
	}, nil)

	// Register all tools. Each registerXxx call wires one tool with
	// schema-inferred input. See tools.go.
	registerAllTools(server, client)

	// stdio transport: the AI client spawns us as a subprocess and
	// talks to us over stdin/stdout. Signal handling shuts us down
	// gracefully if the parent dies (Claude Desktop quit, Cursor
	// reload, etc.).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "lifeastro-mcp: server exited: %v\n", err)
		os.Exit(1)
	}
}
