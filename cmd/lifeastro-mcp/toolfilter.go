package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LIFEASTRO_MCP_TOOLS limits the registered tools. The full list is about
// 312 KB of schemas, more than many clients can take. The value is a
// comma-separated list of tool-name prefixes: a family such as "panchang"
// (every panchang_* tool) or a whole tool name such as "sade_sati". Unset or
// empty registers every tool.

// parseToolFilter splits the LIFEASTRO_MCP_TOOLS value into lower-case
// prefixes, skipping empty entries.
func parseToolFilter(value string) []string {
	var prefixes []string
	for _, p := range strings.Split(value, ",") {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			prefixes = append(prefixes, p)
		}
	}
	return prefixes
}

// toolAllowed reports whether a tool name starts with one of the prefixes.
func toolAllowed(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// limitTools removes every registered tool that no prefix matches and returns
// how many are kept and how many there were. With no prefixes, or when
// nothing matches (a typo must not leave an empty server), all tools stay.
func limitTools(ctx context.Context, server *mcp.Server, prefixes []string) (kept, total int, err error) {
	if len(prefixes) == 0 {
		return 0, 0, nil
	}
	// The server has no public tool listing, so list them as a client would,
	// over an in-memory session that is closed again before stdio starts.
	clientT, serverT := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverT, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("list tools: %w", err)
	}
	defer serverSession.Close()
	session, err := mcp.NewClient(&mcp.Implementation{Name: "lifeastro-mcp-filter", Version: version}, nil).
		Connect(ctx, clientT, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("list tools: %w", err)
	}
	defer session.Close()

	var drop []string
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			return 0, 0, fmt.Errorf("list tools: %w", err)
		}
		total++
		if !toolAllowed(tool.Name, prefixes) {
			drop = append(drop, tool.Name)
		}
	}
	if len(drop) == total {
		return total, total, nil
	}
	server.RemoveTools(drop...)
	return total - len(drop), total, nil
}
