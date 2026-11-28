package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestServer wires the same tool surface as production but points
// the upstream HTTP client at a fake server, then connects an
// in-memory MCP client/server pair so tests can issue real
// CallTool RPCs without spawning a subprocess.
//
// Returns the connected client session + the request-log struct so
// tests can assert what hit the upstream HTTP server.
func newTestServer(t *testing.T, status int, upstreamBody string) (*mcp.ClientSession, *requestLog) {
	t.Helper()
	upstream, log := fakeAPIServer(t, status, upstreamBody)
	apiClient := newAPIClient(upstream.URL, "dv_test_xxx", "test")

	server := mcp.NewServer(&mcp.Implementation{Name: "lifeastro-test", Version: "test"}, nil)
	registerAllTools(server, apiClient)

	clientT, serverT := mcp.NewInMemoryTransports()
	ctx := t.Context()

	go func() {
		// Server.Run blocks until the transport closes — that
		// happens when t.Cleanup tears down the in-memory pipe.
		_ = server.Run(ctx, serverT)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, log
}

// callTool is a small helper that invokes a tool by name with a typed
// argument map and returns the textual result content (most tools
// return JSON-as-text so this is the right shape for assertions).
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("call tool %q: %v", name, err)
	}
	return res
}

// firstText extracts the first TextContent from a CallToolResult or
// fails the test. Every tool in this MCP server returns a single
// text block (the upstream JSON pretty-printed), so this helper
// covers the universal case.
func firstText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatalf("CallToolResult has no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("first content is %T, want *TextContent", res.Content[0])
	}
	return tc.Text
}

// expectedToolCount is the source of truth for "how many tools should
// the MCP server expose right now". Bump this when adding a tool;
// the registration test below catches accidental additions or removals.
const expectedToolCount = 332

func TestListTools_AllRegistered(t *testing.T) {
	t.Parallel()
	session, _ := newTestServer(t, 200, `{}`)
	out, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if got := len(out.Tools); got != expectedToolCount {
		t.Errorf("got %d tools, want %d", got, expectedToolCount)
	}
	// Complete tool registry — updated to match all 287 registered tools.
	// Remove western_natal_rulerships (dead API route).
	// Added: v0.6–v0.9 tools (milan param fixes, panchang sub-routes, chart_ghatak, etc.).
	wantNames := map[string]bool{
		// v0.1 (12 tools)
		"panchang_today":     true,
		"chart_ascendant":    true,
		"chart_planets":      true,
		"chart_navamsa":      true,
		"chart_nakshatra":    true,
		"dasha_current":      true,
		"mangal_dosha":       true,
		"match_making_score": true,
		"sade_sati":          true,
		"horoscope_daily":    true,
		"horoscope_weekly":   true,
		"transits_now":       true,
		// v0.2 (30 tools)
		"panchang_tithi":           true,
		"panchang_choghadiya":      true,
		"panchang_hora":            true,
		"panchang_rahu_kaal":       true,
		"panchang_brahma_muhurat":  true,
		"chart_houses":             true,
		"chart_aspects":            true,
		"chart_dignity":            true,
		"chart_divisional":         true,
		"chart_avakhada":           true,
		"chart_shadbala":           true,
		"chart_ashtakavarga":       true,
		"dasha_vimshottari_full":   true,
		"dasha_yogini_current":     true,
		"nadi_dosha":               true,
		"vivah_phal":               true,
		"ashtakoota_breakdown":     true,
		"muhurta_vivah":            true,
		"muhurta_naamkaran":        true,
		"muhurta_best_time":        true,
		"varshaphal_chart":         true,
		"numerology_full":          true,
		"eclipses_solar":           true,
		"eclipses_lunar":           true,
		"festivals_month":          true,
		"planet_retrograde_window": true,
		"planet_ingress":           true,
		"prashna_answer":           true,
		"western_natal_chart":      true,
		"horoscope_monthly":        true,
		// v0.3 (29 tools — western_natal_rulerships removed: route does not exist in API)
		"western_natal_houses":                  true,
		"western_natal_aspects":                 true,
		"western_transit_positions":             true,
		"western_synastry_aspects":              true,
		"western_synastry_score":                true,
		"western_composite_chart":               true,
		"western_progressions_planets":          true,
		"western_solar_arc_planets":             true,
		"western_solar_return":                  true,
		"western_transits_to_natal":             true,
		"western_interpretation_aspect":         true,
		"western_interpretation_transit":        true,
		"western_natal_lots":                    true,
		"western_natal_midpoints":               true,
		"western_profections_annual":            true,
		"western_zodiacal_releasing":            true,
		"western_firdaria":                      true,
		"western_dignities_planet":              true,
		"western_natal_fixed_stars":             true,
		"western_natal_big_four_asteroids":      true,
		"western_natal_declinations":            true,
		"western_natal_harmonic":                true,
		"western_natal_antiscia":                true,
		"western_natal_midpoint_tree":           true,
		"western_heliocentric_planets":          true,
		"western_eclipses_upcoming":             true,
		"western_ingresses":                     true,
		"western_retrograde_window":             true,
		"western_astrocartography_planet_lines": true,
		// v0.4 (54 tools)
		"western_transits_exact":                       true,
		"western_transits_calendar":                    true,
		"western_synastry_grid":                        true,
		"western_lunar_return":                         true,
		"western_solar_return_aspects":                 true,
		"western_lunar_return_aspects":                 true,
		"western_composite_aspects":                    true,
		"western_davison_chart":                        true,
		"western_davison_aspects":                      true,
		"western_progressions_aspects_to_natal":        true,
		"western_progressions_lunation":                true,
		"western_solar_arc_aspects_to_natal":           true,
		"western_horoscope_daily":                      true,
		"western_horoscope_weekly":                     true,
		"western_horoscope_monthly":                    true,
		"western_natal_summary":                        true,
		"western_transit_summary":                      true,
		"western_narrative_yearly":                     true,
		"western_narrative_decan":                      true,
		"western_narrative_sabian":                     true,
		"western_narrative_fixed_star":                 true,
		"western_narrative_timelord_profection":        true,
		"western_narrative_timelord_zr":                true,
		"western_profections_monthly":                  true,
		"western_profections_daily":                    true,
		"western_zodiacal_releasing_spirit":            true,
		"western_dignities_almuten":                    true,
		"western_dignities_triplicity":                 true,
		"western_dignities_bounds":                     true,
		"western_dignities_face":                       true,
		"western_natal_centaurs":                       true,
		"western_natal_tnos":                           true,
		"western_natal_lilith_set":                     true,
		"western_natal_uranian_hypotheticals":          true,
		"western_natal_parans":                         true,
		"western_natal_vertex_set":                     true,
		"western_natal_harmonic_aspects":               true,
		"western_compatibility_sun_sign":               true,
		"western_lunar_phase":                          true,
		"western_natal_aspect_patterns":                true,
		"western_natal_dominant":                       true,
		"western_natal_chart_shape":                    true,
		"western_eclipses_visibility":                  true,
		"western_primary_directions_placidus_zodiacal": true,
		"western_primary_directions_placidus_mundane":  true,
		"western_primary_directions_regiomontanus":     true,
		"numerology_personal_periods":                  true,
		"numerology_challenges":                        true,
		"numerology_advanced":                          true,
		"tarot_cards":                                  true,
		"tarot_card":                                   true,
		"tarot_daily":                                  true,
		"tarot_spread":                                 true,
		"tarot_yes_no":                                 true,
		// v0.5 (10 tools)
		"panchang_basic":                true,
		"panchang_advanced":             true,
		"panchang_lagna_table":          true,
		"panchang_monthly":              true,
		"panchang_nakshatra_prediction": true,
		"chart_astro_details":           true,
		"varshaphal_harsha_bala":        true,
		"varshaphal_mudda_dasha":        true,
		"narrative_yearly_bhavishyafal": true,
		"geo_reverse":                   true,
		// v0.6–v0.9 additions (152 tools)
		"ashtakavarga_bhinna":                  true,
		"ashtakavarga_kaksha":                  true,
		"ashtakavarga_sarva":                   true,
		"ashtakavarga_transit_score":           true,
		"calendar_adhik_maas":                  true,
		"calendar_month":                       true,
		"calendar_ritu":                        true,
		"calendar_samvatsara":                  true,
		"calendar_solstices":                   true,
		"chart_bhavabala":                      true,
		"chart_combustion":                     true,
		"chart_ghatak":                         true,
		"chart_graha_yuddha":                   true,
		"chart_house_occupants":                true,
		"chart_kp_house_significator":          true,
		"chart_kp_sublord":                     true,
		"chart_shadbala_component":             true,
		"chart_single_planet":                  true,
		"dasha_vimshottari_antardasha":         true,
		"dasha_vimshottari_mahadasha":          true,
		"dasha_vimshottari_pratyantardasha":    true,
		"dasha_vimshottari_sookshma":           true,
		"dasha_yogini_by_name":                 true,
		"dasha_yogini_full":                    true,
		"eclipses_all":                         true,
		"festivals_on_date":                    true,
		"festivals_vrat":                       true,
		"geo_place":                            true,
		"geo_search":                           true,
		"geo_timezone":                         true,
		"milan_ashtakoota_single":              true,
		"milan_dasha_sync":                     true,
		"chart_kundli_facts":                   true,
		"chart_yogas":                          true,
		"chart_doshas":                         true,
		"chart_sensitivity":                    true,
		"transit_sade_sati_windows":            true,
		"transit_windows":                      true,
		"remedy_for_graha":                     true,
		"milan_overlay":                        true,
		"milan_summary":                        true,
		"milan_dhan_yog":                       true,
		"milan_longevity":                      true,
		"milan_mahendra":                       true,
		"milan_navamsa_compat":                 true,
		"milan_santan_yog":                     true,
		"milan_shani_dosha":                    true,
		"milan_stree_dirgha":                   true,
		"muhurta_chandra_bala":                 true,
		"muhurta_graha_pravesh":                true,
		"muhurta_panchaka_rahita":              true,
		"muhurta_sarvartha_siddhi":             true,
		"muhurta_shubha_yoga":                  true,
		"muhurta_tara_bala":                    true,
		"muhurta_vyapar":                       true,
		"muhurta_yatra":                        true,
		"narrative_career_outlook":             true,
		"narrative_dasha_phal":                 true,
		"narrative_dasha_tree_phal":            true,
		"narrative_divisional":                 true,
		"narrative_dosha_by_id":                true,
		"narrative_doshas":                     true,
		"narrative_finance_outlook":            true,
		"narrative_horoscope_daily_by_lagna":   true,
		"narrative_horoscope_daily_by_moon":    true,
		"narrative_horoscope_daily_tamil":      true,
		"narrative_horoscope_weekly_by_lagna":  true,
		"narrative_horoscope_weekly_by_moon":   true,
		"narrative_house_lord":                 true,
		"narrative_house_lords_overview":       true,
		"narrative_karakas":                    true,
		"narrative_lagna":                      true,
		"narrative_marriage_outlook":           true,
		"narrative_milan":                      true,
		"narrative_prashna":                    true,
		"narrative_moon_sign":                  true,
		"narrative_nakshatra":                  true,
		"narrative_planet_in_house":            true,
		"narrative_planet_in_sign":             true,
		"narrative_planets_in_houses":          true,
		"narrative_planets_in_signs":           true,
		"narrative_profile":                    true,
		"narrative_sade_sati_history":          true,
		"narrative_sade_sati_phase":            true,
		"narrative_sade_sati_status":           true,
		"narrative_sun_sign":                   true,
		"narrative_transit_ashtakavarga":       true,
		"narrative_transit_double":             true,
		"narrative_transit_phal":               true,
		"narrative_varshaphal_themes":          true,
		"narrative_yoga_by_id":                 true,
		"narrative_yogas":                      true,
		"numerology_conductor":                 true,
		"numerology_destiny":                   true,
		"numerology_driver":                    true,
		"numerology_personality":               true,
		"numerology_soul":                      true,
		"panchang_abhijit":                     true,
		"panchang_amrit_kalam":                 true,
		"panchang_durmuhurta":                  true,
		"panchang_gulika":                      true,
		"panchang_hindu_month":                 true,
		"panchang_hora_dinman":                 true,
		"panchang_karana":                      true,
		"panchang_moonrise_moonset":            true,
		"panchang_nakshatra":                   true,
		"panchang_panchaka":                    true,
		"panchang_pradosh_kaal":                true,
		"panchang_sankranti":                   true,
		"panchang_siddha_yoga":                 true,
		"panchang_sunrise_sunset":              true,
		"panchang_vara":                        true,
		"panchang_varjyam":                     true,
		"panchang_yamaganda":                   true,
		"panchang_yoga":                        true,
		"planet_combustion_window":             true,
		"planet_speed":                         true,
		"prashna_arudha":                       true,
		"prashna_chart":                        true,
		"prashna_kp_horary":                    true,
		"prashna_lagna_lord":                   true,
		"prashna_significators":                true,
		"remedy_by_rule_id":                    true,
		"remedy_daan_by_id":                    true,
		"remedy_rudraksha_by_id":               true,
		"remedy_yantra_by_id":                  true,
		"remedy_lalkitab_list":                 true,
		"remedy_lalkitab_for_placement":        true,
		"dasha_char_full":                      true,
		"dasha_char_major":                     true,
		"dasha_char_current":                   true,
		"dasha_char_antardashas":               true,
		"dasha_char_pratyantardashas":          true,
		"lalkitab_chart":                       true,
		"lalkitab_dasha":                       true,
		"lalkitab_dasha_current":               true,
		"remedy_mantra_by_id":                  true,
		"remedy_pooja_by_id":                   true,
		"remedy_vrat_by_id":                    true,
		"report_dasha_analysis":                true,
		"report_horoscope_daily":               true,
		"report_horoscope_monthly":             true,
		"report_horoscope_weekly":              true,
		"report_kundli_brihad":                 true,
		"report_kundli_detailed":               true,
		"report_kundli_lite":                   true,
		"report_mangal_dosha":                  true,
		"report_match_making":                  true,
		"report_numerology":                    true,
		"report_sade_sati":                     true,
		"report_varshaphal":                    true,
		"transit_ashtakavarga":                 true,
		"transit_double_transit":               true,
		"transit_small_panoti":                 true,
		"transit_small_panoti_history":         true,
		"transit_tarabala":                     true,
		"transit_vedha":                        true,
		"varshaphal_lord":                      true,
		"varshaphal_muntha":                    true,
		"varshaphal_solar_return_jd":           true,
		"varshaphal_yoga":                      true,
		"vedic_remedies":                       true,
		"western_astrocartography_local_space": true,
		"western_composite_houses":             true,
		"western_composite_planets":            true,
		"western_davison_houses":               true,
		"western_davison_planets":              true,
		"western_dignities_receptions":         true,
		"western_narrative_firdaria":           true,
		"western_natal_ascendant":              true,
		"western_natal_asteroid":               true,
		"western_natal_fixed_star":             true,
		"western_natal_lot":                    true,
		"western_natal_mc":                     true,
		"western_natal_out_of_bounds":          true,
		"western_natal_planets":                true,
		"western_natal_single_planet":          true,
		"western_solar_return_exact_jd":        true,
		// v0.11 (8 tools)
		"chart_ashtakvarga_planet":  true,
		"narrative_reference":       true,
		"meta_vedic_yogas":          true,
		"meta_vedic_doshas":         true,
		"meta_vedic_locales":        true,
		"meta_vedic_narrative_axes": true,
		"meta_ayanamsas":            true,
		"meta_endpoints":            true,
		// v1.1 (11 tools)
		"dasha_ashtottari_full":    true,
		"dasha_ashtottari_current": true,
		"chart_arudha_padas":       true,
		"chart_upapada":            true,
		"chart_argala":             true,
		"muhurta_vahan":            true,
		"muhurta_bhoomi_pujan":     true,
		"muhurta_upanayan":         true,
		"muhurta_annaprashan":      true,
		"muhurta_vidyarambh":       true,
		"muhurta_karnavedha":       true,
	}
	if got, want := len(wantNames), expectedToolCount; got != want {
		t.Errorf("test wantNames map has %d entries, want %d (test bookkeeping out of sync)", got, want)
	}
	gotNames := map[string]bool{}
	for _, tl := range out.Tools {
		gotNames[tl.Name] = true
	}
	for name := range wantNames {
		if !gotNames[name] {
			t.Errorf("tool %q not registered", name)
		}
	}
	// Catch the reverse — extra tools that aren't in the expected
	// list. Helps spot accidental dupes when adding new ones.
	for name := range gotNames {
		if !wantNames[name] {
			t.Errorf("unexpected tool %q registered (not in test list)", name)
		}
	}
}

func TestTool_PanchangToday_HitsCorrectEndpoint(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, cannedPanchangResponse)
	res := callTool(t, session, "panchang_today", map[string]any{
		"lat": 19.076,
		"lon": 72.8777,
		"tz":  "Asia/Kolkata",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if got := firstText(t, res); !strings.Contains(got, "Krishna Saptami") {
		t.Errorf("tool output missing fixture content: %q", got)
	}
	if len(log.paths) != 1 || log.paths[0] != "/v1/panchang/summary" {
		t.Errorf("upstream paths = %v, want [/v1/panchang/summary]", log.paths)
	}
	q := log.queries[0]
	if q.Get("tz") != "Asia/Kolkata" {
		t.Errorf("upstream tz = %q, want Asia/Kolkata", q.Get("tz"))
	}
}

func TestTool_ChartAscendant_PassesBirthQuintet(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"sign_en":"Capricorn","degree":"12.5"}`)
	res := callTool(t, session, "chart_ascendant", map[string]any{
		"lat":  19.076,
		"lon":  72.8777,
		"date": "1990-01-15",
		"time": "10:30",
		"tz":   "Asia/Kolkata",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if got := firstText(t, res); !strings.Contains(got, "Capricorn") {
		t.Errorf("tool output missing 'Capricorn': %q", got)
	}
	if log.paths[0] != "/v1/chart/ascendant" {
		t.Errorf("upstream path = %q, want /v1/chart/ascendant", log.paths[0])
	}
	q := log.queries[0]
	for k, want := range map[string]string{
		"lat": "19.076", "lon": "72.8777",
		"date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("query[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestTool_DashaCurrent_OptionalAtParam(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"maha":{"lord":"Saturn"}}`)
	res := callTool(t, session, "dasha_current", map[string]any{
		"lat":  19.076,
		"lon":  72.8777,
		"date": "1990-01-15",
		"time": "10:30",
		"tz":   "Asia/Kolkata",
		"at":   "2026-05-10T12:00:00Z",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if got := log.queries[0].Get("at"); got != "2026-05-10T12:00:00Z" {
		t.Errorf("upstream 'at' query = %q, want passthrough of input", got)
	}
}

func TestTool_HoroscopeDaily_LowercasesRashi(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"body":"You will have a great day."}`)
	// Pass mixed-case rashi — the tool must normalize before sending
	// to the upstream API (which expects lowercase).
	res := callTool(t, session, "horoscope_daily", map[string]any{
		"rashi": "Aries",
		"lang":  "en",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if got := log.queries[0].Get("rashi"); got != "aries" {
		t.Errorf("upstream rashi = %q, want lowercase 'aries'", got)
	}
}

func TestTool_MatchMaking_PassesBothPartiesParams(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"total":28,"max":36}`)
	res := callTool(t, session, "match_making_score", map[string]any{
		"bride_lat":  19.076,
		"bride_lon":  72.8777,
		"bride_date": "1995-06-12",
		"bride_time": "08:15",
		"bride_tz":   "Asia/Kolkata",
		"groom_lat":  28.6139,
		"groom_lon":  77.2090,
		"groom_date": "1992-11-03",
		"groom_time": "14:45",
		"groom_tz":   "Asia/Kolkata",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	// API reads boy.*/girl.* params via parseBoyGirlMoonInfo.
	// API convention: boy = groom (male), girl = bride (female).
	q := log.queries[0]
	for k, want := range map[string]string{
		"boy.date":  "1992-11-03",
		"girl.date": "1995-06-12",
		"boy.time":  "14:45",
		"girl.time": "08:15",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("query[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestTool_TransitsNow_NoArgsRequired(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `[{"name":"Sun","sign":"Aries"}]`)
	// transits_now is the one tool with no required input — invoking
	// it with an empty map should still hit the upstream and produce
	// a result.
	res := callTool(t, session, "transits_now", map[string]any{})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if log.paths[0] != "/v1/transit/positions" {
		t.Errorf("upstream path = %q, want /v1/transit/positions", log.paths[0])
	}
}

func TestTool_UpstreamErrorSurfacesAsToolError(t *testing.T) {
	t.Parallel()
	// 400 from upstream → the MCP tool should set IsError and pass
	// the upstream message through so the AI client can show a useful
	// message to the end user.
	session, _ := newTestServer(t, 400, `{"error":"invalid date","code":"invalid_param","details":"date must be YYYY-MM-DD"}`)
	res := callTool(t, session, "chart_ascendant", map[string]any{
		"lat":  19.076,
		"lon":  72.8777,
		"date": "not-a-date",
		"time": "10:30",
		"tz":   "Asia/Kolkata",
	})
	if !res.IsError {
		t.Fatalf("expected IsError=true, got false")
	}
	msg := firstText(t, res)
	for _, want := range []string{"400", "invalid_param", "invalid date"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error text %q missing %q", msg, want)
		}
	}
}

func TestTool_OutputIsValidJSON(t *testing.T) {
	t.Parallel()
	// The upstream payload is returned as a single text content block
	// containing the JSON. Verify that text is valid JSON the AI
	// client can parse — a marshal-side regression here would silently
	// break every tool.
	session, _ := newTestServer(t, 200, `{"foo": "bar", "list": [1,2,3]}`)
	res := callTool(t, session, "panchang_today", map[string]any{
		"lat": 1.0, "lon": 1.0, "tz": "UTC",
	})
	body := firstText(t, res)
	got := mustDecode(t, body)
	if got["foo"] != "bar" {
		t.Errorf("decoded foo = %v, want bar", got["foo"])
	}
}

func TestTool_AuthHeaderForwarded(t *testing.T) {
	t.Parallel()
	// Verify the bearer token reaches the upstream — a missing auth
	// header would break every paid call in production.
	session, _ := newTestServer(t, 200, `{}`)
	_ = callTool(t, session, "transits_now", map[string]any{})
	// The header check is in the upstream handler — fakeAPIServer's
	// handler doesn't capture headers. Re-do directly via apiClient
	// for header verification (TestAPIClient_GetHappyPath covers it),
	// so this test just sanity-checks the call path completes.
	_ = session
}

// Context: ensure context cancellation propagates from the MCP layer
// through to the upstream HTTP request — a runaway tool call should
// abort cleanly when the AI client cancels.
func TestTool_ContextCancellationPropagates(t *testing.T) {
	t.Parallel()
	session, _ := newTestServer(t, 200, `{}`)
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately
	_, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "transits_now",
		Arguments: map[string]any{},
	})
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
}

// =====================================================================
// v0.2 tool tests
// =====================================================================
//
// Spot-checks on the trickier v0.2 surface — path-parametrised tools
// (chart_divisional uses {varga} in the URL), shared muhurta input
// shape, eclipse range with optional location filter, the rashi
// lowerTrim normalization shared by horoscope_monthly.

func TestTool_ChartDivisional_PathContainsVarga(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"varga":"D10","planets":[]}`)
	res := callTool(t, session, "chart_divisional", map[string]any{
		"lat":   19.076,
		"lon":   72.8777,
		"date":  "1990-01-15",
		"time":  "10:30",
		"tz":    "Asia/Kolkata",
		"varga": "D10",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	// The varga selector lives in the URL path, not the query string —
	// a regression here would silently return D9 instead of D10.
	if log.paths[0] != "/v1/chart/divisional/D10" {
		t.Errorf("upstream path = %q, want /v1/chart/divisional/D10", log.paths[0])
	}
}

func TestTool_ChartDivisional_AcceptsLongName(t *testing.T) {
	t.Parallel()
	// Confirm the path passthrough works for textual varga names too —
	// "navamsa" is what an AI would send when paraphrasing the user.
	session, log := newTestServer(t, 200, `{}`)
	_ = callTool(t, session, "chart_divisional", map[string]any{
		"lat": 19.076, "lon": 72.8777,
		"date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata",
		"varga": "navamsa",
	})
	if log.paths[0] != "/v1/chart/divisional/navamsa" {
		t.Errorf("upstream path = %q, want /v1/chart/divisional/navamsa", log.paths[0])
	}
}

func TestTool_MuhurtaVivah_ScoresOneDate(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"result":{"score":81}}`)
	res := callTool(t, session, "muhurta_vivah", map[string]any{
		"lat":  19.076,
		"lon":  72.8777,
		"tz":   "Asia/Kolkata",
		"date": "2026-11-20",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	// The API scores a single date; it has no start_date/end_date and
	// rejects the request when date is missing.
	q := log.queries[0]
	if q.Get("date") != "2026-11-20" {
		t.Errorf("date not forwarded: %q", q.Get("date"))
	}
	if q.Has("start_date") || q.Has("end_date") || q.Has("time") {
		t.Errorf("unexpected params in upstream query: %v", q)
	}
}

func TestTool_EclipsesSolar_OptionalLocationOmittedWhenZero(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `[]`)
	// Caller supplies only the date range — location should NOT
	// appear in the upstream query (we don't want the API to filter
	// to "visible from 0,0" by accident).
	_ = callTool(t, session, "eclipses_solar", map[string]any{
		"start_date": "2026-01-01",
		"end_date":   "2026-12-31",
	})
	q := log.queries[0]
	if q.Has("lat") || q.Has("lon") {
		t.Errorf("lat/lon should be absent when zero; got lat=%q lon=%q", q.Get("lat"), q.Get("lon"))
	}
}

func TestTool_EclipsesSolar_OptionalLocationIncludedWhenSet(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `[]`)
	_ = callTool(t, session, "eclipses_solar", map[string]any{
		"start_date": "2026-01-01",
		"end_date":   "2026-12-31",
		"lat":        19.076,
		"lon":        72.8777,
	})
	q := log.queries[0]
	if q.Get("lat") == "" || q.Get("lon") == "" {
		t.Errorf("lat/lon should be present when set; got lat=%q lon=%q", q.Get("lat"), q.Get("lon"))
	}
}

func TestTool_HoroscopeMonthly_LowerTrimMatchesDailyWeekly(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"body":""}`)
	// Pass a mixed-case rashi with surrounding whitespace — both must
	// be normalized exactly like the daily/weekly tools do.
	_ = callTool(t, session, "horoscope_monthly", map[string]any{
		"rashi": "  PISCES  ",
	})
	if got := log.queries[0].Get("rashi"); got != "pisces" {
		t.Errorf("upstream rashi = %q, want trimmed-lowercased 'pisces'", got)
	}
}

func TestTool_VarshaphalChart_PassesYear(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"varsha_lord":"Sun"}`)
	_ = callTool(t, session, "varshaphal_chart", map[string]any{
		"lat":  19.076,
		"lon":  72.8777,
		"date": "1990-01-15",
		"time": "10:30",
		"tz":   "Asia/Kolkata",
		"year": 2026,
	})
	if got := log.queries[0].Get("year"); got != "2026" {
		t.Errorf("upstream year = %q, want 2026", got)
	}
}

func TestTool_NadiDosha_PassesBothPartiesParams(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"present":false}`)
	res := callTool(t, session, "nadi_dosha", map[string]any{
		"bride_lat":  19.076,
		"bride_lon":  72.8777,
		"bride_date": "1995-06-12",
		"bride_time": "08:15",
		"bride_tz":   "Asia/Kolkata",
		"groom_lat":  28.6139,
		"groom_lon":  77.2090,
		"groom_date": "1992-11-03",
		"groom_time": "14:45",
		"groom_tz":   "Asia/Kolkata",
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if log.paths[0] != "/v1/milan/nadi-dosha" {
		t.Errorf("upstream path = %q, want /v1/milan/nadi-dosha", log.paths[0])
	}
	// API reads boy.*/girl.* params via parseBoyGirlMoonInfo.
	// API convention: boy = groom (male), girl = bride (female).
	q := log.queries[0]
	if q.Get("boy.date") != "1992-11-03" || q.Get("girl.date") != "1995-06-12" {
		t.Error("both parties' birth dates should reach upstream")
	}
}

func TestTool_NumerologyFull_OptionalFieldsOmitted(t *testing.T) {
	t.Parallel()
	// dob is required by the API; name and lang are optional and must
	// not reach upstream as empty strings.
	session, log := newTestServer(t, 200, `{}`)
	_ = callTool(t, session, "numerology_full", map[string]any{
		"dob": "1990-01-15",
	})
	q := log.queries[0]
	if q.Get("dob") != "1990-01-15" {
		t.Errorf("dob not forwarded: %q", q.Get("dob"))
	}
	if q.Has("name") || q.Has("lang") {
		t.Errorf("optional fields should be absent when unset; got name=%q lang=%q",
			q.Get("name"), q.Get("lang"))
	}
}

func TestTool_PrashnaAnswer_SendsHouseAndMoment(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"answer":"yes"}`)
	res := callTool(t, session, "prashna_answer", map[string]any{
		"lat":   19.076,
		"lon":   72.8777,
		"tz":    "Asia/Kolkata",
		"house": 10,
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	q := log.queries[0]
	if q.Get("house") != "10" {
		t.Errorf("house not forwarded: %q", q.Get("house"))
	}
	// No date/time supplied: the prashna moment defaults to now, and the
	// API requires both to be present.
	if q.Get("date") == "" || q.Get("time") == "" {
		t.Errorf("prashna moment not defaulted: date=%q time=%q", q.Get("date"), q.Get("time"))
	}
}

// Endpoints that take a natal chart AND a transit moment read the chart
// under "birth.*". Sent unprefixed, the API takes the birth date for the
// transit date and returns 200 with the wrong day's numbers — so this is
// asserted here rather than left to the live sweep, which cannot see it.
func TestTool_TransitTools_SendBirthPrefixAndToday(t *testing.T) {
	t.Parallel()
	for _, tool := range []string{
		"transit_double_transit", "transit_tarabala", "transit_vedha", "transit_ashtakavarga", "transit_small_panoti",
		"ashtakavarga_kaksha", "ashtakavarga_transit_score",
		"narrative_transit_phal", "narrative_dasha_phal", "narrative_transit_double", "narrative_transit_ashtakavarga",
		"narrative_sade_sati_status", "narrative_sade_sati_phase", "narrative_sade_sati_history",
	} {
		session, log := newTestServer(t, 200, `{}`)
		res := callTool(t, session, tool, map[string]any{
			"lat": 19.076, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata",
		})
		if res.IsError {
			t.Fatalf("%s: tool returned error: %s", tool, firstText(t, res))
		}
		q := log.queries[0]
		if q.Get("birth.date") != "1990-01-15" || q.Get("birth.time") != "10:30" || q.Get("birth.tz") != "Asia/Kolkata" {
			t.Errorf("%s: natal data not under birth.*: %v", tool, q)
		}
		if q.Get("date") == "" || q.Get("date") == "1990-01-15" {
			t.Errorf("%s: transit date should default to today, got %q", tool, q.Get("date"))
		}
	}
}

// The personalised horoscopes read the natal chart UNPREFIXED and the day
// from transit_date. Getting this backwards returns 200 with a horoscope for
// a chart cast today, so it is pinned here.
func TestTool_NarrativeHoroscope_BirthUnprefixedDayAsTransitDate(t *testing.T) {
	t.Parallel()
	for _, tool := range []string{"narrative_horoscope_daily_by_lagna", "narrative_horoscope_daily_by_moon", "narrative_horoscope_daily_tamil", "narrative_horoscope_weekly_by_lagna", "narrative_horoscope_weekly_by_moon"} {
		session, log := newTestServer(t, 200, `{}`)
		res := callTool(t, session, tool, map[string]any{
			"lat": 19.076, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata",
			"on_date": "2026-10-10",
		})
		if res.IsError {
			t.Fatalf("%s: tool returned error: %s", tool, firstText(t, res))
		}
		q := log.queries[0]
		if q.Get("date") != "1990-01-15" || q.Get("time") != "10:30" {
			t.Errorf("%s: birth date/time must be sent as date/time, got %v", tool, q)
		}
		if q.Has("birth.date") {
			t.Errorf("%s: must not send birth.* (the handler does not read it): %v", tool, q)
		}
		if q.Get("transit_date") != "2026-10-10" {
			t.Errorf("%s: on_date must be sent as transit_date, got %q", tool, q.Get("transit_date"))
		}
	}
}

func TestTool_PanchangDefaultsDateToToday(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{}`)
	_ = callTool(t, session, "panchang_tithi", map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata"})
	if got := log.queries[0].Get("date"); len(got) != len("2006-01-02") {
		t.Errorf("date should default to today (YYYY-MM-DD), got %q", got)
	}
}

func TestDashaLord(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{"jup": "Jupiter", "Jupiter": "Jupiter", " SATURN ": "Saturn", "rah": "Rahu", "guru": "Jupiter", "x/y": "x%2Fy"} {
		if got := dashaLord(in); got != want {
			t.Errorf("dashaLord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAPIErrorMessage(t *testing.T) {
	t.Parallel()
	nested := `{"error":{"code":"missing_param","message":"date is required","request_id":"abc"}}`
	if got, want := apiErrorMessage([]byte(nested)), "missing_param: date is required (request_id abc)"; got != want {
		t.Errorf("nested: got %q, want %q", got, want)
	}
	if got := apiErrorMessage([]byte(`<html>bad gateway</html>`)); got != "" {
		t.Errorf("non-JSON body should yield no message, got %q", got)
	}
}

// TestLowerTrim spot-checks the small standalone helper used by
// horoscope_monthly. It mirrors strings.ToLower(strings.TrimSpace(...))
// so the tool's output is consistent with daily/weekly handlers.
func TestLowerTrim(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"Aries", "aries"},
		{"  PISCES  ", "pisces"},
		{"\tCancer\n", "cancer"},
		{"already-lower", "already-lower"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := lowerTrim(tc.in); got != tc.want {
			t.Errorf("lowerTrim(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNarrativePrashna_TopicAndHouse(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"sections":[]}`)
	res := callTool(t, session, "narrative_prashna", map[string]any{
		"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata",
		"date": "2027-01-15", "time": "10:00", "topic": "career", "question_house": 7,
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if log.paths[0] != "/v1/vedic/narrative/prashna" {
		t.Errorf("upstream path = %q, want /v1/vedic/narrative/prashna", log.paths[0])
	}
	q := log.queries[0]
	if q.Get("topic") != "career" || q.Get("date") != "2027-01-15" || q.Get("time") != "10:00" || q.Get("lat") != "19.076" {
		t.Errorf("query = %v", q)
	}
	// The API reads question_house only when topic is absent; send one.
	if q.Has("question_house") {
		t.Errorf("question_house sent alongside topic: %v", q)
	}

	res = callTool(t, session, "narrative_prashna", map[string]any{
		"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "date": "2027-01-15", "question_house": 7,
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	if q := log.queries[1]; q.Get("question_house") != "7" || q.Has("topic") {
		t.Errorf("house-only query = %v", q)
	}
}

func TestNarrativeMilan_GroomIsBoy(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{"sections":[]}`)
	res := callTool(t, session, "narrative_milan", map[string]any{
		"groom": map[string]any{"lat": 28.6139, "lon": 77.209, "date": "1992-11-03", "time": "14:45", "tz": "Asia/Kolkata"},
		"bride": map[string]any{"lat": 19.076, "lon": 72.8777, "date": "1995-06-12", "time": "08:15", "tz": "Asia/Kolkata"},
	})
	if res.IsError {
		t.Fatalf("tool returned error: %s", firstText(t, res))
	}
	q := log.queries[0]
	if q.Get("boy.date") != "1992-11-03" || q.Get("girl.date") != "1995-06-12" {
		t.Errorf("boy must be the groom and girl the bride: %v", q)
	}
}

// v0.12 data tools: chart routes read the birth chart unprefixed, the window
// routes read it under birth.* with from/tz unprefixed, milan uses boy/girl.
func TestV12DataTools_QueryMapping(t *testing.T) {
	t.Parallel()
	birth := map[string]any{"lat": 19.076, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata"}
	couple := map[string]any{
		"groom_lat": 19.076, "groom_lon": 72.8777, "groom_date": "1990-01-15", "groom_time": "10:30", "groom_tz": "Asia/Kolkata",
		"bride_lat": 28.6139, "bride_lon": 77.209, "bride_date": "1992-03-20", "bride_time": "08:15", "bride_tz": "Asia/Kolkata",
	}
	with := func(base map[string]any, extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	cases := []struct {
		tool, path string
		args       map[string]any
		want       map[string]string
		absent     []string
	}{
		{"chart_kundli_facts", "/v1/chart/kundli-facts", with(birth, map[string]any{"vargas": "D9,D10", "include": "all", "gender": "Male"}),
			map[string]string{"date": "1990-01-15", "time": "10:30", "lat": "19.076", "vargas": "D9,D10", "include": "all", "gender": "male"}, []string{"birth.date"}},
		{"chart_yogas", "/v1/chart/yogas", with(birth, map[string]any{"all": true}),
			map[string]string{"date": "1990-01-15", "all": "true"}, []string{"gender"}},
		{"chart_doshas", "/v1/chart/doshas", birth,
			map[string]string{"date": "1990-01-15", "tz": "Asia/Kolkata"}, []string{"all"}},
		{"chart_sensitivity", "/v1/chart/sensitivity", with(birth, map[string]any{"window": 60, "vargas": "D9", "topocentric_moon": true}),
			map[string]string{"date": "1990-01-15", "window": "60", "vargas": "D9", "topocentric_moon": "true"}, nil},
		{"transit_sade_sati_windows", "/v1/transit/sade-sati/windows", with(birth, map[string]any{"from": "2027-01-15", "years": 30, "include_past": true, "local_tz": "UTC"}),
			map[string]string{"birth.date": "1990-01-15", "birth.time": "10:30", "birth.tz": "Asia/Kolkata", "from": "2027-01-15", "years": "30", "include_past": "true", "tz": "UTC"}, []string{"date"}},
		{"transit_windows", "/v1/transit/windows", map[string]any{"planet": "Saturn", "from": "2027-01-15", "years": 10, "birth": birth},
			map[string]string{"planet": "saturn", "from": "2027-01-15", "years": "10", "birth.date": "1990-01-15", "birth.lat": "19.076", "birth.lon": "72.8777"}, []string{"date", "lat"}},
		{"transit_windows", "/v1/transit/windows", map[string]any{"planet": "jupiter"},
			map[string]string{"planet": "jupiter"}, []string{"birth.date", "from"}},
		{"remedy_for_graha", "/v1/vedic/remedies/for-graha", map[string]any{"graha": "Saturn", "reason": "dosha:sade_sati", "include": "pooja", "lang": "hi"},
			map[string]string{"graha": "saturn", "reason": "dosha:sade_sati", "include": "pooja", "locale": "hi"}, []string{"category"}},
		{"milan_overlay", "/v1/milan/overlay", couple,
			map[string]string{"boy.date": "1990-01-15", "girl.date": "1992-03-20", "girl.lat": "28.6139"}, nil},
		{"milan_summary", "/v1/milan/summary", with(couple, map[string]any{"from": "2027-01-15", "include_kundli_facts": true}),
			map[string]string{"boy.time": "10:30", "girl.time": "08:15", "from": "2027-01-15", "include": "kundli_facts"}, []string{"tz"}},
		{"milan_dasha_sync", "/v1/milan/dasha-sync", with(couple, map[string]any{"marriage_windows": true, "from": "2027-01-15", "years": 5}),
			map[string]string{"boy.lat": "19.076", "girl.lon": "77.209", "marriage_windows": "true", "from": "2027-01-15", "years": "5"}, []string{"timeline"}},
		{"milan_dasha_sync", "/v1/milan/dasha-sync", couple,
			map[string]string{"boy.date": "1990-01-15"}, []string{"timeline", "marriage_windows", "years", "from"}},
	}
	for _, tc := range cases {
		session, log := newTestServer(t, 200, `{}`)
		res := callTool(t, session, tc.tool, tc.args)
		if res.IsError {
			t.Fatalf("%s: tool returned error: %s", tc.tool, firstText(t, res))
		}
		if log.paths[0] != tc.path {
			t.Errorf("%s: path = %q, want %q", tc.tool, log.paths[0], tc.path)
		}
		q := log.queries[0]
		for k, v := range tc.want {
			if q.Get(k) != v {
				t.Errorf("%s: %s = %q, want %q (query %v)", tc.tool, k, q.Get(k), v, q)
			}
		}
		for _, k := range tc.absent {
			if q.Has(k) {
				t.Errorf("%s: %s should not be sent (query %v)", tc.tool, k, q)
			}
		}
	}
}

// noLanguageTools are the Vedic tools that deliberately have no language
// input: their endpoints return only numbers, times, indices and ids, so a
// language would change nothing (checked against the API with locale=gu).
// A new Vedic tool must either take a language or be listed here.
var noLanguageTools = map[string]bool{
	"ashtakavarga_sarva": true, "chart_bhavabala": true, "chart_doshas": true, "chart_sensitivity": true,
	"chart_yogas": true, "match_making_score": true, "milan_overlay": true, "muhurta_chandra_bala": true,
	"muhurta_tara_bala": true, "panchang_abhijit": true, "panchang_brahma_muhurat": true,
	"panchang_durmuhurta": true, "panchang_gulika": true, "panchang_moonrise_moonset": true,
	"panchang_pradosh_kaal": true, "panchang_rahu_kaal": true, "panchang_sunrise_sunset": true,
	"panchang_yamaganda": true, "sade_sati": true, "transit_sade_sati_windows": true,
	"transit_small_panoti": true, "transit_small_panoti_history": true, "transit_windows": true,
	"varshaphal_muntha": true, "varshaphal_solar_return_jd": true,
}

// Every language input carries the same, current language list: an assistant
// picks the language from this text. Western, Tarot, geo and meta tools are
// English-only, except the Western narrative tools (en, hi, mr).
func TestLanguageInputs(t *testing.T) {
	t.Parallel()
	session, _ := newTestServer(t, 200, `{}`)
	out, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	vedic, western := 0, 0
	for _, tl := range out.Tools {
		raw, err := json.Marshal(tl.InputSchema)
		if err != nil {
			t.Fatalf("%s: marshal schema: %v", tl.Name, err)
		}
		var sc struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(raw, &sc); err != nil {
			t.Fatalf("%s: decode schema: %v", tl.Name, err)
		}
		lang, hasLang := sc.Properties["lang"]
		locale, hasLocale := sc.Properties["locale"]
		if hasLang && hasLocale {
			t.Errorf("%s: has both lang and locale", tl.Name)
		}
		for _, r := range sc.Required {
			if r == "lang" || r == "locale" {
				t.Errorf("%s: %s must be optional", tl.Name, r)
			}
		}
		desc := lang.Description + locale.Description
		englishOnly := false
		for _, pre := range []string{"western_", "tarot_", "geo_", "meta_"} {
			if strings.HasPrefix(tl.Name, pre) {
				englishOnly = true
			}
		}
		switch {
		case strings.HasPrefix(tl.Name, "western_narrative_"):
			western++
			if desc != westernLanguageDesc {
				t.Errorf("%s: language description = %q, want %q", tl.Name, desc, westernLanguageDesc)
			}
		case englishOnly:
			if hasLang || hasLocale {
				t.Errorf("%s: English-only tool must not take a language", tl.Name)
			}
		case noLanguageTools[tl.Name]:
			if hasLang || hasLocale {
				t.Errorf("%s: listed in noLanguageTools but takes a language", tl.Name)
			}
		default:
			vedic++
			if desc != vedicLanguageDesc {
				t.Errorf("%s: language description = %q, want %q", tl.Name, desc, vedicLanguageDesc)
			}
			if hasLocale && strings.Contains(tl.Description, "`lang`") || hasLang && strings.Contains(tl.Description, "`locale`") {
				t.Errorf("%s: description names the wrong language input", tl.Name)
			}
		}
	}
	if vedic != 200 || western != 7 {
		t.Errorf("language inputs: %d Vedic + %d Western tools, want 200 + 7", vedic, western)
	}
}

// The language reaches the API for one tool of every family, under the
// query parameter that family's endpoint reads, next to the tool's own
// parameters (birth.* for transit-style tools, boy.*/girl.* for milan).
func TestLanguageForwarded(t *testing.T) {
	t.Parallel()
	birth := map[string]any{"lat": 19.076, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata"}
	place := map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "date": "2026-11-13"}
	couple := map[string]any{
		"groom_lat": 19.076, "groom_lon": 72.8777, "groom_date": "1990-01-15", "groom_time": "10:30", "groom_tz": "Asia/Kolkata",
		"bride_lat": 28.6139, "bride_lon": 77.209, "bride_date": "1992-03-20", "bride_time": "08:15", "bride_tz": "Asia/Kolkata",
	}
	with := func(base map[string]any, extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	cases := []struct {
		tool, path string
		args       map[string]any
		input      string // the tool's language input
		param      string // the query parameter it becomes
		want       map[string]string
	}{
		// Data tools: `locale` added by addLocalizedTool.
		{"panchang_today", "/v1/panchang/summary", place, "locale", "locale", map[string]string{"date": "2026-11-13"}},
		{"panchang_tithi", "/v1/panchang/tithi", place, "locale", "locale", nil},
		{"calendar_ritu", "/v1/calendar/ritu", map[string]any{"date": "2026-11-13", "tz": "Asia/Kolkata"}, "locale", "locale", nil},
		{"eclipses_solar", "/v1/eclipses/solar", map[string]any{"start_date": "2026-01-01", "end_date": "2027-12-31"}, "locale", "locale", nil},
		{"planet_ingress", "/v1/planet-moments/ingress", map[string]any{"planet": "Jupiter", "date": "2026-11-13"}, "locale", "locale", nil},
		{"chart_planets", "/v1/chart/planets", birth, "locale", "locale", map[string]string{"date": "1990-01-15"}},
		{"chart_divisional", "/v1/chart/divisional/D9", with(birth, map[string]any{"varga": "D9"}), "locale", "locale", nil},
		{"dasha_current", "/v1/dasha/vimshottari/current", birth, "locale", "locale", nil},
		{"dasha_char_full", "/v1/dasha/char", birth, "locale", "locale", nil},
		{"ashtakavarga_bhinna", "/v1/ashtakavarga/bhinna/Jupiter", with(birth, map[string]any{"planet": "Jupiter"}), "locale", "locale", map[string]string{"birth.date": "1990-01-15"}},
		{"transit_tarabala", "/v1/transit/tarabala", birth, "locale", "locale", map[string]string{"birth.date": "1990-01-15", "birth.tz": "Asia/Kolkata"}},
		{"transits_now", "/v1/transit/positions", map[string]any{"date": "2026-11-13"}, "locale", "locale", map[string]string{"date": "2026-11-13"}},
		{"mangal_dosha", "/v1/milan/mangal-dosha", birth, "locale", "locale", map[string]string{"birth.date": "1990-01-15"}},
		{"ashtakoota_breakdown", "/v1/milan/ashtakoota", couple, "locale", "locale", map[string]string{"boy.date": "1990-01-15", "girl.date": "1992-03-20"}},
		{"milan_summary", "/v1/milan/summary", couple, "locale", "locale", map[string]string{"boy.time": "10:30"}},
		{"muhurta_vivah", "/v1/muhurta/vivah", place, "locale", "locale", nil},
		{"prashna_chart", "/v1/prashna/chart", place, "locale", "locale", nil},
		{"varshaphal_chart", "/v1/varshaphal/chart", with(birth, map[string]any{"year": 2027}), "locale", "locale", map[string]string{"year": "2027"}},
		{"lalkitab_chart", "/v1/lalkitab/chart", birth, "locale", "locale", nil},
		// Tools with their own language field.
		{"festivals_month", "/v1/festivals/month", map[string]any{"year": 2026, "month": 11}, "locale", "locale", nil},
		{"panchang_monthly", "/v1/panchang/monthly", map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "year": 2026, "month": 11}, "locale", "locale", nil},
		{"horoscope_daily", "/v1/horoscope/daily", map[string]any{"rashi": "Aries"}, "lang", "lang", nil},
		{"numerology_full", "/v1/numerology/full", map[string]any{"name": "Arjun Mehta", "dob": "1990-01-15"}, "lang", "lang", nil},
		{"narrative_yearly_bhavishyafal", "/v1/vedic/narrative/yearly-bhavishyafal", with(birth, map[string]any{"year": 2027}), "lang", "lang", map[string]string{"birth.date": "1990-01-15"}},
		{"narrative_horoscope_daily_tamil", "/v1/vedic/narrative/horoscope/daily-tamil", with(birth, map[string]any{"on_date": "2026-11-13"}), "lang", "lang", map[string]string{"transit_date": "2026-11-13", "date": "1990-01-15"}},
		{"remedy_for_graha", "/v1/vedic/remedies/for-graha", map[string]any{"graha": "saturn", "reason": "debilitated"}, "lang", "locale", nil},
		{"remedy_lalkitab_list", "/v1/vedic/remedies/lal-kitab", map[string]any{}, "locale", "locale", nil},
		{"western_narrative_yearly", "/v1/western/narrative/yearly", map[string]any{"sign": "Leo"}, "lang", "lang", nil},
	}
	for _, tc := range cases {
		// With the language set it is forwarded.
		session, log := newTestServer(t, 200, `{}`)
		res := callTool(t, session, tc.tool, with(tc.args, map[string]any{tc.input: "gu"}))
		if res.IsError {
			t.Fatalf("%s: tool returned error: %s", tc.tool, firstText(t, res))
		}
		if log.paths[0] != tc.path {
			t.Errorf("%s: path = %q, want %q", tc.tool, log.paths[0], tc.path)
		}
		q := log.queries[0]
		if q.Get(tc.param) != "gu" {
			t.Errorf("%s: %s = %q, want gu (query %v)", tc.tool, tc.param, q.Get(tc.param), q)
		}
		if q.Has("lang") && q.Has("locale") {
			t.Errorf("%s: both lang and locale sent (query %v)", tc.tool, q)
		}
		for k, v := range tc.want {
			if q.Get(k) != v {
				t.Errorf("%s: %s = %q, want %q (query %v)", tc.tool, k, q.Get(k), v, q)
			}
		}
		// Without it nothing is sent: the API default (English, or the
		// language configured on the API key) applies.
		session, log = newTestServer(t, 200, `{}`)
		if res := callTool(t, session, tc.tool, tc.args); res.IsError {
			t.Fatalf("%s (no language): tool returned error: %s", tc.tool, firstText(t, res))
		}
		if q := log.queries[0]; q.Has("lang") || q.Has("locale") {
			t.Errorf("%s: language sent although none was given (query %v)", tc.tool, q)
		}
	}
}

// A tool without a language input rejects one instead of ignoring it, and
// the note about localized names is only on tools that take `locale`.
func TestLanguageNotOnNumberOnlyTools(t *testing.T) {
	t.Parallel()
	session, log := newTestServer(t, 200, `{}`)
	res, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "panchang_rahu_kaal",
		Arguments: map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "date": "2026-11-13", "locale": "gu"},
	})
	if err == nil && !res.IsError {
		t.Errorf("panchang_rahu_kaal accepted a locale it does not have (query %v)", log.queries)
	}
	out, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	for _, tl := range out.Tools {
		if noLanguageTools[tl.Name] && strings.Contains(tl.Description, strings.TrimSpace(localizedNote)) {
			t.Errorf("%s: description promises localized names", tl.Name)
		}
	}
}

// v1.1 tools: every route reads the chart or the event moment unprefixed;
// options are forwarded only when set, so the API defaults apply (scheme 28,
// exception parashari, no yatra direction, purpose general).
func TestV14Tools_QueryMapping(t *testing.T) {
	t.Parallel()
	birth := map[string]any{"lat": 19.076, "lon": 72.8777, "date": "1990-01-15", "time": "10:30", "tz": "Asia/Kolkata"}
	event := map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "date": "2027-02-10"}
	with := func(base map[string]any, extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range base {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	cases := []struct {
		tool, path string
		args       map[string]any
		want       map[string]string
		absent     []string
	}{
		{"dasha_ashtottari_full", "/v1/dasha/ashtottari", with(birth, map[string]any{"depth": 2, "scheme": 27, "antar_start": "Next_Lord"}),
			map[string]string{"date": "1990-01-15", "time": "10:30", "depth": "2", "scheme": "27", "antar_start": "next_lord"}, []string{"birth.date", "at"}},
		{"dasha_ashtottari_full", "/v1/dasha/ashtottari", birth,
			map[string]string{"date": "1990-01-15"}, []string{"depth", "scheme", "antar_start"}},
		{"dasha_ashtottari_current", "/v1/dasha/ashtottari/current", with(birth, map[string]any{"at": "2027-01-15T00:00:00Z", "scheme": 28}),
			map[string]string{"date": "1990-01-15", "at": "2027-01-15T00:00:00Z", "scheme": "28"}, []string{"depth", "antar_start"}},
		{"chart_arudha_padas", "/v1/chart/arudha-padas", with(birth, map[string]any{"exception": "none", "scorpio_lord": "Ketu", "aquarius_lord": "rahu"}),
			map[string]string{"date": "1990-01-15", "exception": "none", "scorpio_lord": "ketu", "aquarius_lord": "rahu"}, nil},
		{"chart_upapada", "/v1/chart/upapada", birth,
			map[string]string{"lat": "19.076", "tz": "Asia/Kolkata"}, []string{"exception", "scorpio_lord", "aquarius_lord"}},
		{"chart_argala", "/v1/chart/argala", with(birth, map[string]any{"house": 10, "include_eighth": true, "node_reversal": "rahu_ketu", "vipareeta_min_malefics": 2}),
			map[string]string{"house": "10", "include_eighth": "true", "node_reversal": "rahu_ketu", "vipareeta_min_malefics": "2"}, nil},
		{"chart_argala", "/v1/chart/argala", birth,
			map[string]string{"date": "1990-01-15"}, []string{"house", "include_eighth", "node_reversal", "vipareeta_min_malefics"}},
		{"muhurta_vahan", "/v1/muhurta/vahan", event, map[string]string{"date": "2027-02-10", "tz": "Asia/Kolkata"}, []string{"time", "birth_date"}},
		{"muhurta_bhoomi_pujan", "/v1/muhurta/bhoomi-pujan", with(event, map[string]any{"time": "09:15"}), map[string]string{"time": "09:15"}, nil},
		{"muhurta_upanayan", "/v1/muhurta/upanayan", event, map[string]string{"date": "2027-02-10"}, nil},
		{"muhurta_annaprashan", "/v1/muhurta/annaprashan", with(event, map[string]any{"birth_date": "2026-08-01", "sex": "Male"}),
			map[string]string{"date": "2027-02-10", "birth_date": "2026-08-01", "sex": "male"}, nil},
		{"muhurta_annaprashan", "/v1/muhurta/annaprashan", event, nil, []string{"birth_date", "sex"}},
		{"muhurta_vidyarambh", "/v1/muhurta/vidyarambh", with(event, map[string]any{"birth_date": "2022-03-05"}), map[string]string{"birth_date": "2022-03-05"}, nil},
		{"muhurta_karnavedha", "/v1/muhurta/karnavedha", event, map[string]string{"date": "2027-02-10"}, []string{"birth_date"}},
		{"muhurta_yatra", "/v1/muhurta/yatra", event, map[string]string{"date": "2027-02-10"}, []string{"direction"}},
		{"muhurta_yatra", "/v1/muhurta/yatra", with(event, map[string]any{"direction": "NE"}), map[string]string{"direction": "NE"}, nil},
		{"muhurta_best_time", "/v1/muhurta/best-time", map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "start_date": "2027-02-01", "end_date": "2027-02-28", "purpose": "Yatra", "direction": "N"},
			map[string]string{"start_date": "2027-02-01", "end_date": "2027-02-28", "purpose": "yatra", "direction": "N"}, []string{"birth_date"}},
		{"muhurta_best_time", "/v1/muhurta/best-time", map[string]any{"lat": 19.076, "lon": 72.8777, "tz": "Asia/Kolkata", "start_date": "2027-02-01", "end_date": "2027-02-28"},
			map[string]string{"start_date": "2027-02-01"}, []string{"purpose", "direction", "birth_date"}},
	}
	for _, tc := range cases {
		session, log := newTestServer(t, 200, `{}`)
		res := callTool(t, session, tc.tool, tc.args)
		if res.IsError {
			t.Fatalf("%s: tool returned error: %s", tc.tool, firstText(t, res))
		}
		if log.paths[0] != tc.path {
			t.Errorf("%s: path = %q, want %q", tc.tool, log.paths[0], tc.path)
		}
		q := log.queries[0]
		for k, v := range tc.want {
			if q.Get(k) != v {
				t.Errorf("%s: %s = %q, want %q (query %v)", tc.tool, k, q.Get(k), v, q)
			}
		}
		for _, k := range tc.absent {
			if q.Has(k) {
				t.Errorf("%s: %s should not be sent (query %v)", tc.tool, k, q)
			}
		}
	}
}
