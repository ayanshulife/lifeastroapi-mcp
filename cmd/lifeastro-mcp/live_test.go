//go:build live

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Live compatibility sweep: calls EVERY registered tool against the real
// API and fails on any tool the API rejects. This is the check that the
// tool surface (paths AND parameter names/values) matches production.
//
// Not part of the normal test run — it needs network + a real key and
// spends credits (one request per tool):
//
//	LIFEASTRO_API_KEY=dv_live_... go test -tags live -run TestLive ./cmd/lifeastro-mcp/ -v
//
// LIFEASTRO_LIVE_DUMP=1 prints every tool's input properties instead of
// calling anything (used to maintain liveArgs below).

type liveSchema struct {
	Properties map[string]struct {
		Type        any    `json:"type"`
		Description string `json:"description"`
	} `json:"properties"`
	Required []string `json:"required"`
}

func newLiveSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	key := os.Getenv("LIFEASTRO_API_KEY")
	if key == "" {
		t.Skip("LIFEASTRO_API_KEY not set")
	}
	base := os.Getenv("LIFEASTRO_API_URL")
	if base == "" {
		base = defaultAPIBaseURL
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "lifeastro-live", Version: "live"}, nil)
	registerAllTools(server, newAPIClient(base, key, "live-test"))
	clientT, serverT := mcp.NewInMemoryTransports()
	ctx := t.Context()
	go func() { _ = server.Run(ctx, serverT) }()
	client := mcp.NewClient(&mcp.Implementation{Name: "live-client", Version: "live"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func liveTools(t *testing.T, s *mcp.ClientSession) []*mcp.Tool {
	t.Helper()
	var all []*mcp.Tool
	cursor := ""
	for {
		out, err := s.ListTools(t.Context(), &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		all = append(all, out.Tools...)
		if out.NextCursor == "" {
			break
		}
		cursor = out.NextCursor
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name < all[j].Name })
	return all
}

func schemaOf(t *testing.T, tl *mcp.Tool) liveSchema {
	t.Helper()
	raw, err := json.Marshal(tl.InputSchema)
	if err != nil {
		t.Fatalf("%s: marshal schema: %v", tl.Name, err)
	}
	var sc liveSchema
	if err := json.Unmarshal(raw, &sc); err != nil {
		t.Fatalf("%s: decode schema: %v", tl.Name, err)
	}
	return sc
}

func TestLiveDumpInputs(t *testing.T) {
	if os.Getenv("LIFEASTRO_LIVE_DUMP") == "" {
		t.Skip("set LIFEASTRO_LIVE_DUMP=1 to dump")
	}
	s := newLiveSession(t)
	for _, tl := range liveTools(t, s) {
		sc := schemaOf(t, tl)
		req := map[string]bool{}
		for _, r := range sc.Required {
			req[r] = true
		}
		names := make([]string, 0, len(sc.Properties))
		for n := range sc.Properties {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			p := sc.Properties[n]
			mark := "opt"
			if req[n] {
				mark = "REQ"
			}
			fmt.Printf("DUMP\t%s\t%s\t%s\t%v\t%s\n", tl.Name, n, mark, p.Type, p.Description)
		}
	}
}

func TestLiveAllTools(t *testing.T) {
	s := newLiveSession(t)
	tools := liveTools(t, s)

	type result struct{ name, msg string }
	var (
		mu     sync.Mutex
		failed []result
		wg     sync.WaitGroup
		sem    = make(chan struct{}, 3)
		// ~2 requests/second keeps a full sweep under the API's
		// per-minute rate limit.
		pace = time.NewTicker(500 * time.Millisecond)
	)
	defer pace.Stop()
	only := os.Getenv("LIFEASTRO_LIVE_ONLY") // comma-separated tool names to re-run
	// Name prefixes to leave out, e.g. "report_" — the composed reports
	// cost 500–8000 credits each, so a routine sweep skips them.
	skip := strings.Split(os.Getenv("LIFEASTRO_LIVE_SKIP"), ",")
	for _, tl := range tools {
		if only != "" && !strings.Contains(","+only+",", ","+tl.Name+",") {
			continue
		}
		skipped := false
		for _, pre := range skip {
			if pre != "" && strings.HasPrefix(tl.Name, pre) {
				skipped = true
			}
		}
		if skipped {
			continue
		}
		sc := schemaOf(t, tl)
		args := map[string]any{}
		var missing []string
		for _, n := range sc.Required {
			v, ok := liveArg(tl.Name, n)
			if !ok {
				missing = append(missing, n)
				continue
			}
			args[n] = v
		}
		// Optional args that a tool-specific override names explicitly.
		for n, v := range liveToolArgs[tl.Name] {
			args[n] = v
		}
		if len(missing) > 0 {
			failed = append(failed, result{tl.Name, "no sample value for required: " + strings.Join(missing, ",")})
			continue
		}
		<-pace.C
		wg.Add(1)
		go func(name string, args map[string]any) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			res, err := s.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
			msg := ""
			switch {
			case err != nil:
				msg = "rpc error: " + err.Error()
			case res.IsError:
				msg = "tool error"
				if len(res.Content) > 0 {
					if tc, ok := res.Content[0].(*mcp.TextContent); ok {
						msg = tc.Text
					}
				}
			}
			if msg != "" {
				if len(msg) > 400 {
					msg = msg[:400]
				}
				a, _ := json.Marshal(args)
				mu.Lock()
				failed = append(failed, result{name, msg + "  args=" + string(a)})
				mu.Unlock()
			}
		}(tl.Name, args)
	}
	wg.Wait()

	sort.Slice(failed, func(i, j int) bool { return failed[i].name < failed[j].name })
	for _, f := range failed {
		t.Errorf("FAIL %s: %s", f.name, f.msg)
	}
	t.Logf("live sweep: %d tools, %d failed", len(tools), len(failed))
}

// liveArg returns a realistic sample value for a required input property.
// Tool-specific values win over the by-name defaults; person-prefixed
// properties (a_/b_/groom_/bride_/birth_) fall back to a birth sample.
func liveArg(tool, prop string) (any, bool) {
	if m, ok := liveToolArgs[tool]; ok {
		if v, ok := m[prop]; ok {
			return v, true
		}
	}
	if v, ok := liveDefaults[prop]; ok {
		return v, true
	}
	for _, pre := range []string{"a_", "groom_", "birth_"} {
		if rest, ok := strings.CutPrefix(prop, pre); ok {
			v, ok := livePersonA[rest]
			return v, ok
		}
	}
	for _, pre := range []string{"b_", "bride_"} {
		if rest, ok := strings.CutPrefix(prop, pre); ok {
			v, ok := livePersonB[rest]
			return v, ok
		}
	}
	return nil, false
}

var livePersonA = map[string]any{"lat": 19.0760, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata"}
var livePersonB = map[string]any{"lat": 28.6139, "lon": 77.2090, "date": "1992-06-20", "time": "14:15", "tz": "Asia/Kolkata"}

// liveDefaults: sample values keyed by property name (a 1990 Mumbai birth).
var liveDefaults = map[string]any{
	"lat": 19.0760, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata",
	"year": 2026, "month": 11, "start_date": "2026-11-01", "end_date": "2026-11-30",
	"planet": "Jupiter", "rashi": "Aries", "sign": "Leo", "question": "Will I get the job?",
	"dob": "1990-01-15", "birth_date": "1990-01-15", "age_in_years": 35.5, "house": 5,
	"aspect": "trine", "asc_sign": "leo", "varga": "D9", "transit_planet": "Saturn",
	"natal_planet": "Sun", "return_year": 2026, "md": "Jupiter", "ad": "Saturn", "pd": "Mercury", "sd": "Ketu",
	"lot_sign": "cancer", "lord": "Jupiter", "birth_nak": 7, "year_month": "2026-11",
	"week_start": "2026-11-02", "tz_offset_hours": 5.5, "star": "Regulus",
	"sign_a": "Aries", "sign_b": "Leo", "q": "Mumbai", "planet_a": "Sun", "planet_b": "Moon",
	"observer_lat": 28.6139, "observer_lon": 77.2090, "number": 1, "n": 5, "maha": "Jupiter",
	"koota": "nadi", "degree": 100, "decan": 2, "cusp": 7, "component": "sthana",
	"birth_moon_sign": 3, "asc_lon": 123.4,
	"groom": livePersonA, "bride": livePersonB,
}

// liveToolArgs: per-tool sample values where the by-name default is wrong
// or the property is tool-specific.
var liveToolArgs = map[string]map[string]any{
	"chart_single_planet":           {"name": "Jupiter"},
	"dasha_yogini_by_name":          {"name": "mangala"},
	"numerology_advanced":           {"name": "Vikas Sharma"},
	"numerology_destiny":            {"name": "Vikas Sharma"},
	"numerology_personality":        {"name": "Vikas Sharma"},
	"numerology_soul":               {"name": "Vikas Sharma"},
	"report_numerology":             {"name": "Vikas Sharma"},
	"western_natal_fixed_star":      {"name": "Regulus"},
	"western_natal_lot":             {"name": "fortune"},
	"western_natal_single_planet":   {"name": "Venus"},
	"dasha_char_antardashas":        {"md": "leo"},
	"dasha_char_pratyantardashas":   {"md": "leo", "ad": "scorpio"},
	"western_retrograde_window":     {"planet": "Mercury"},
	"remedy_lalkitab_for_placement": {"planet": "saturn", "house": "8"},
	"narrative_dosha_by_id":         {"id": "mangal_dosha"},
	"narrative_yoga_by_id":          {"id": "hamsa"},
	"tarot_card":                    {"id": 5},
	"tarot_spread":                  {"type": "three"},
	"western_lunar_return":          {"from": "2026-06-01T00:00:00Z"},
	"western_lunar_return_aspects":  {"from": "2026-06-01T00:00:00Z"},
	"western_transits_calendar":     {"from": "2026-11-01", "to": "2026-11-20"},
	"western_transits_exact":        {"from": "2026-01-01T00:00:00Z", "to": "2027-01-01T00:00:00Z"},
	"eclipses_all":                  {"start_date": "2026-01-01", "end_date": "2027-12-31"},
	"eclipses_solar":                {"start_date": "2026-01-01", "end_date": "2027-12-31"},
	"eclipses_lunar":                {"start_date": "2026-01-01", "end_date": "2027-12-31"},
	"muhurta_naamkaran":             {"date": "2026-11-20", "birth_date": "2026-11-05"},
	"narrative_reference":           {"family": "yoga"},
	"geo_place":                     {"id": "1275339"},
	"remedy_by_rule_id":             {"id": "hanuman_chalisa"},
	"remedy_mantra_by_id":           {"id": "hanuman_chalisa"},
	"remedy_daan_by_id":             {"id": "sadesati_setting_shanti"},
	"remedy_vrat_by_id":             {"id": "kaalsarp_dosha_shanti"},
	"remedy_pooja_by_id":            {"id": "jupiter_dasha_shanti"},
	"remedy_yantra_by_id":           {"id": "vedic.L1.remedy.yantra.jupiter"},
	"remedy_rudraksha_by_id":        {"id": "vedic.L1.remedy.rudraksha.jupiter"},
}
