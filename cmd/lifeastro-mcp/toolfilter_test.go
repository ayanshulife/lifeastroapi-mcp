package main

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func filteredToolNames(t *testing.T, filter string) (names []string, kept, total int) {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "lifeastro-test", Version: "test"}, nil)
	registerAllTools(server, newAPIClient("http://127.0.0.1:0", "dv_test_xxx", "test"))
	kept, total, err := limitTools(t.Context(), server, parseToolFilter(filter))
	if err != nil {
		t.Fatalf("limitTools: %v", err)
	}

	clientT, serverT := mcp.NewInMemoryTransports()
	go func() { _ = server.Run(t.Context(), serverT) }()
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil).
		Connect(t.Context(), clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	for tool, err := range session.Tools(t.Context(), nil) {
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		names = append(names, tool.Name)
	}
	return names, kept, total
}

func TestToolFilter_KeepsOnlyMatchingPrefixes(t *testing.T) {
	t.Parallel()
	names, kept, total := filteredToolNames(t, " Panchang, sade_sati ,,")
	if total != expectedToolCount || kept != len(names) || kept == 0 || kept >= total {
		t.Fatalf("kept %d of %d, listed %d", kept, total, len(names))
	}
	sawSadeSati := false
	for _, n := range names {
		if !strings.HasPrefix(n, "panchang") && n != "sade_sati" {
			t.Errorf("tool %q should have been removed", n)
		}
		sawSadeSati = sawSadeSati || n == "sade_sati"
	}
	if !sawSadeSati {
		t.Error("sade_sati should be listed")
	}
}

func TestToolFilter_UnsetOrNoMatchKeepsAll(t *testing.T) {
	t.Parallel()
	for _, filter := range []string{"", " , ", "no_such_family"} {
		if names, _, _ := filteredToolNames(t, filter); len(names) != expectedToolCount {
			t.Errorf("filter %q: %d tools listed, want %d", filter, len(names), expectedToolCount)
		}
	}
}

// The v1.1 tools belong to the existing families, so a family filter keeps them.
func TestToolFilter_NewToolsInFamilies(t *testing.T) {
	t.Parallel()
	names, _, _ := filteredToolNames(t, "dasha,chart,muhurta")
	got := map[string]bool{}
	for _, n := range names {
		got[n] = true
	}
	for _, n := range []string{
		"dasha_ashtottari_full", "dasha_ashtottari_current", "chart_arudha_padas", "chart_upapada", "chart_argala",
		"muhurta_vahan", "muhurta_bhoomi_pujan", "muhurta_upanayan", "muhurta_annaprashan", "muhurta_vidyarambh", "muhurta_karnavedha",
	} {
		if !got[n] {
			t.Errorf("filter dasha,chart,muhurta dropped %s", n)
		}
	}
}
