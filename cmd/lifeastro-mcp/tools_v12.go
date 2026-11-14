package main

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerV12Tools adds the generic chart / transit / remedy / milan data
// endpoints (2026-11). They return ids and numbers only, no prose:
// planet ids sun..ketu, sign_index 0-based (Aries = 0), house 1-based
// whole-sign from the lagna, nakshatra index 0-based (Ashwini = 0), every
// instant as *_local (RFC 3339) + *_jd (Julian Day, UT), and a skipped
// object listing anything that could not be computed.
//
//	Chart (4):    /v1/chart/kundli-facts, /yogas, /doshas, /sensitivity
//	Transit (2):  /v1/transit/sade-sati/windows, /v1/transit/windows
//	Remedies (1): /v1/vedic/remedies/for-graha
//	Milan (2):    /v1/milan/overlay, /v1/milan/summary
//
// /v1/milan/dasha-sync gained optional timeline params; milan_dasha_sync
// (tools_v06.go) takes them.
//
// Total new: 9 tools.
func registerV12Tools(s *mcp.Server, c *apiClient) {
	registerChartKundliFacts(s, c)
	registerChartYogas(s, c)
	registerChartDoshas(s, c)
	registerChartSensitivity(s, c)
	registerTransitSadeSatiWindows(s, c)
	registerTransitWindows(s, c)
	registerRemedyForGraha(s, c)
	registerMilanOverlay(s, c)
	registerMilanSummary(s, c)
}

// idsNote is appended to every tool description in this file.
const idsNote = " Data only, ids not names: planets sun..ketu, sign_index 0-based (0 = Aries), house 1-based whole-sign from the lagna, nakshatra index 0-based (0 = Ashwini); map them to names in the user's language."

func setIf(q url.Values, key, v string) {
	if v = strings.TrimSpace(v); v != "" {
		q.Set(key, v)
	}
}

func setIntIf(q url.Values, key string, v int) {
	if v != 0 {
		q.Set(key, strconv.Itoa(v))
	}
}

func setTrueIf(q url.Values, key string, v bool) {
	if v {
		q.Set(key, "true")
	}
}

// --------------------------------------------------------------------
// Chart
// --------------------------------------------------------------------

type ChartKundliFactsInput struct {
	BirthInput
	Vargas  string `json:"vargas,omitempty" jsonschema:"optional comma list of divisional charts for planets[].varga, e.g. D9,D10 (D2..D60; default D9; at most 16; D1 is not allowed)"`
	Include string `json:"include,omitempty" jsonschema:"optional comma list of sections: shadbala, bhavabala, ashtakvarga, aspects, avastha, karakas, avakhada, yogas, or all (default: everything except yogas)"`
	Gender  string `json:"gender,omitempty" jsonschema:"optional male or female; only narrows the Mahabhagya yoga when yogas are included"`
}

func registerChartKundliFacts(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "chart_kundli_facts",
		Description: "Get the whole Vedic birth chart as structured data in one call (50 credits): lagna; every graha with sign, degree, house, nakshatra, pada, retrograde, combust, dignity, neecha bhanga, vargottama, houses it rules, shadbala, avastha and divisional placements; every house with sign, lord, SAV bindus, bhava bala, occupants and aspecting grahas; Jaimini chara karakas; avakhada. include=yogas also adds the yoga and dosha items. Best single tool when you need to reason about a chart yourself. Use for 'analyse my kundli', 'which planets are strong in my chart', 'full chart data'." + idsNote,
		Title:       "Kundli Facts (Full Chart Data)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartKundliFactsInput) (*mcp.CallToolResult, any, error) {
		q := in.BirthInput.toQuery()
		setIf(q, "vargas", in.Vargas)
		setIf(q, "include", in.Include)
		setIf(q, "gender", strings.ToLower(in.Gender))
		return callPassthrough(ctx, c, "/v1/chart/kundli-facts", q)
	})
}

type ChartYogaDoshaInput struct {
	BirthInput
	All    bool   `json:"all,omitempty" jsonschema:"optional; true also lists every catalogue id that is absent (present:false). Default: present items plus cancelled ones"`
	Gender string `json:"gender,omitempty" jsonschema:"optional male or female; narrows the Mahabhagya yoga"`
}

func (in ChartYogaDoshaInput) toQuery() url.Values {
	q := in.BirthInput.toQuery()
	setTrueIf(q, "all", in.All)
	setIf(q, "gender", strings.ToLower(in.Gender))
	return q
}

func registerChartYogas(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "chart_yogas",
		Description: "List the yogas in a Vedic birth chart as data, without text (20 credits): each item has id, category, present, the grahas and houses that form it, strength, cancelled + cancel_reasons, sub_ids and the classical source. Includes neecha bhanga per planet and kendra-trikona raja yogas. Pair with narrative_yoga_by_id when the user wants the written meaning of one yoga. Use for 'which yogas are in my chart', 'do I have Gajakesari yoga', 'raja yogas in this kundli'." + idsNote,
		Title:       "Yogas in a Chart (Data)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartYogaDoshaInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/chart/yogas", in.toQuery())
	})
}

func registerChartDoshas(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "chart_doshas",
		Description: "List the doshas in a Vedic birth chart as data, without text (20 credits): Kaal Sarp (with type), Mangal (counted from lagna, Moon and Venus, with cancellation reasons), Pitra, combust, debilitation (with neecha bhanga), retrograde, gandanta, grahan, papakartari, mrityu bhaga, guru chandal, visha kanya and ganda moola. A cancelled dosha comes back with present:false, cancelled:true. Use for 'do I have mangal dosha', 'is kaal sarp dosha in my chart', 'which doshas are in this kundli'." + idsNote,
		Title:       "Doshas in a Chart (Data)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartYogaDoshaInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/chart/doshas", in.toQuery())
	})
}

type ChartSensitivityInput struct {
	BirthInput
	Window          int    `json:"window,omitempty" jsonschema:"optional half-width of the scan in minutes (default 120, max 180)"`
	Vargas          string `json:"vargas,omitempty" jsonschema:"optional comma list of vargas besides D1, e.g. D9,D10,D24 (the default)"`
	TopocentricMoon bool   `json:"topocentric_moon,omitempty" jsonschema:"optional; true uses the topocentric Moon for the nakshatra and pada spans"`
}

func registerChartSensitivity(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "chart_sensitivity",
		Description: "Check how sensitive a birth chart is to the birth time (20 credits): for the D1 lagna and each varga lagna, and for the Moon's nakshatra and pada, how many minutes before and after the given time the value stays the same and what it changes to; plus any graha changing sign inside the window. Use when the birth time is approximate, before trusting D9/D10/D60 readings, or for 'does my lagna change if I was born 10 minutes later', 'how accurate does my birth time need to be'. Times come as *_local and *_jd." + idsNote,
		Title:       "Birth-Time Sensitivity",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartSensitivityInput) (*mcp.CallToolResult, any, error) {
		q := in.BirthInput.toQuery()
		setIntIf(q, "window", in.Window)
		setIf(q, "vargas", in.Vargas)
		setTrueIf(q, "topocentric_moon", in.TopocentricMoon)
		return callPassthrough(ctx, c, "/v1/chart/sensitivity", q)
	})
}

// --------------------------------------------------------------------
// Transit windows
// --------------------------------------------------------------------

type TransitSadeSatiWindowsInput struct {
	BirthInput
	From        string `json:"from,omitempty" jsonschema:"optional range start YYYY-MM-DD (local midnight); defaults to now"`
	Years       int    `json:"years,omitempty" jsonschema:"optional span in years after from (default 25, max 60)"`
	IncludePast bool   `json:"include_past,omitempty" jsonschema:"optional; true starts the range at the birth moment, so past Sade Sati periods are included"`
	LocalTz     string `json:"local_tz,omitempty" jsonschema:"optional IANA zone for the *_local times and for from; defaults to the birth tz"`
}

func registerTransitSadeSatiWindows(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "transit_sade_sati_windows",
		Description: "Get exact dated Sade Sati and Dhaiya (Small Panoti, 4th/8th) windows from the natal Moon for up to 60 years (20 credits): intervals (each stay of Saturn in one sign, kind sade_sati with phase rising/peak/setting, dhaiya_4 kantak or dhaiya_8 ashtama, exact ingress times, retro re-entries), cycles (merged windows with phase spans) and at_from (the status today). Use for 'when is my next Sade Sati', 'exact Sade Sati dates', 'when does my Shani dhaiya end'." + idsNote,
		Title:       "Sade Sati Windows (Dated)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in TransitSadeSatiWindowsInput) (*mcp.CallToolResult, any, error) {
		q := birthPrefixQuery(in.BirthInput)
		setIf(q, "from", in.From)
		setIntIf(q, "years", in.Years)
		setTrueIf(q, "include_past", in.IncludePast)
		setIf(q, "tz", in.LocalTz)
		return callPassthrough(ctx, c, "/v1/transit/sade-sati/windows", q)
	})
}

// TransitWindowsBirth is the optional natal chart for transit_windows.
// date/time/tz give house_from_moon; lat/lon add house_from_lagna.
type TransitWindowsBirth struct {
	Date string   `json:"date" jsonschema:"birth date YYYY-MM-DD"`
	Time string   `json:"time" jsonschema:"birth time HH:MM 24-hour"`
	Tz   string   `json:"tz" jsonschema:"IANA timezone of the birth (e.g. Asia/Kolkata)"`
	Lat  *float64 `json:"lat,omitempty" jsonschema:"optional birth latitude; with lon adds house_from_lagna"`
	Lon  *float64 `json:"lon,omitempty" jsonschema:"optional birth longitude"`
}

type TransitWindowsInput struct {
	Planet string               `json:"planet" jsonschema:"planet id: sun, moon, mars, mercury, jupiter, venus, saturn, rahu or ketu"`
	From   string               `json:"from,omitempty" jsonschema:"optional range start YYYY-MM-DD (local midnight); defaults to now"`
	Years  int                  `json:"years,omitempty" jsonschema:"optional span in years: default/max 25/60 for jupiter, saturn, rahu, ketu; 5/20 for sun, mercury, venus, mars; 1/2 for moon"`
	Tz     string               `json:"tz,omitempty" jsonschema:"optional IANA zone for the *_local times and for from; defaults to the birth tz, else UTC"`
	Birth  *TransitWindowsBirth `json:"birth,omitempty" jsonschema:"optional natal chart; adds house_from_moon (and house_from_lagna with lat/lon) to every window"`
}

func registerTransitWindows(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "transit_windows",
		Description: "List every stay of one planet in a sidereal sign over a range of years, with exact ingress dates, retrograde re-entries and retrograde_at_start (20 credits). With the optional birth chart, each window also gives the house from the natal Moon and lagna. Use for 'when does Jupiter enter my 7th house', 'Saturn transit dates for the next 10 years', 'when is Rahu in Pisces'." + idsNote,
		Title:       "Planet Transit Windows",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in TransitWindowsInput) (*mcp.CallToolResult, any, error) {
		q := url.Values{}
		q.Set("planet", strings.ToLower(strings.TrimSpace(in.Planet)))
		setIf(q, "from", in.From)
		setIntIf(q, "years", in.Years)
		setIf(q, "tz", in.Tz)
		if b := in.Birth; b != nil {
			setIf(q, "birth.date", b.Date)
			setIf(q, "birth.time", b.Time)
			setIf(q, "birth.tz", b.Tz)
			if b.Lat != nil && b.Lon != nil {
				q.Set("birth.lat", strconv.FormatFloat(*b.Lat, 'f', -1, 64))
				q.Set("birth.lon", strconv.FormatFloat(*b.Lon, 'f', -1, 64))
			}
		}
		return callPassthrough(ctx, c, "/v1/transit/windows", q)
	})
}

// --------------------------------------------------------------------
// Remedies
// --------------------------------------------------------------------

type RemedyForGrahaInput struct {
	Graha    string `json:"graha,omitempty" jsonschema:"graha id: sun, moon, mars, mercury, jupiter, venus, saturn, rahu or ketu (required unless reason is dosha:<id>)"`
	Reason   string `json:"reason,omitempty" jsonschema:"optional: debilitated, combust, dusthana, weak_shadbala, dasha, or dosha:<id> such as dosha:sade_sati, dosha:mangal, dosha:kaalsarp, dosha:pitra"`
	Include  string `json:"include,omitempty" jsonschema:"optional comma list of extra kinds beyond the free ones: pooja, pilgrimage, yantra, rudraksha"`
	Category string `json:"category,omitempty" jsonschema:"optional: limit to one catalogue category (mantra, daan, vrat, pooja, yantra, rudraksha)"`
	Lang     string `json:"lang,omitempty" jsonschema:"response language: en (default), hi, mr, te, ta, kn, ml, gu, bn (pa: names only)"`
}

func registerRemedyForGraha(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "remedy_for_graha",
		Description: "Get remedies for one graha by reason (10 credits): mantra, daan, vrat and other remedies from the LifeAstroAPI remedy catalogue, each with kind, weekday (0 = Sunday) and why it matched. reason: debilitated, combust, dusthana, weak_shadbala, dasha (adds the dasha shanti), or dosha:<id> (sade_sati, mangal, kaalsarp, pitra). Free kinds by default; include adds pooja, pilgrimage, yantra, rudraksha. Use after chart_kundli_facts or chart_doshas shows a weak graha or a dosha, or for 'remedies for weak Saturn', 'what to do in Rahu dasha', 'upay for sade sati'.",
		Title:       "Remedies for a Graha",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in RemedyForGrahaInput) (*mcp.CallToolResult, any, error) {
		q := url.Values{}
		setIf(q, "graha", strings.ToLower(in.Graha))
		setIf(q, "reason", strings.ToLower(in.Reason))
		setIf(q, "include", in.Include)
		setIf(q, "category", strings.ToLower(in.Category))
		setIf(q, "locale", in.Lang)
		return callPassthrough(ctx, c, "/v1/vedic/remedies/for-graha", q)
	})
}

// --------------------------------------------------------------------
// Milan
// --------------------------------------------------------------------

func registerMilanOverlay(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "milan_overlay",
		Description: "Overlay a couple's two Vedic charts as data (20 credits): each partner's grahas in the other's houses (with conjunctions), graha pairs in the same sign and the same nakshatra, the Moon sign relation (e.g. 2/12, 6/8) and where each partner's lagna lord sits in the other's chart (house, dignity, combust, retro). Each chart uses the ayanamsa of that partner's birth moment. Use for 'his planets in my chart', 'synastry for our kundlis', 'how do our charts interact'." + idsNote,
		Title:       "Milan Chart Overlay",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in AshtakootaInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/milan/overlay", boyGirlQuery(in))
	})
}

type MilanSummaryInput struct {
	AshtakootaInput
	From               string `json:"from,omitempty" jsonschema:"optional start of the 10-year Sade Sati range, YYYY-MM-DD; defaults to now"`
	LocalTz            string `json:"local_tz,omitempty" jsonschema:"optional IANA zone for from and the Sade Sati *_local times; defaults to each partner's birth tz"`
	IncludeKundliFacts bool   `json:"include_kundli_facts,omitempty" jsonschema:"optional; true also returns both partners' full chart data (chart_kundli_facts)"`
}

func registerMilanSummary(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "milan_summary",
		Description: "Get all the kundli matching data for a couple in one call (60 credits), with numbers identical to the separate tools: Ashtakoota with every koota, navamsa compatibility, Mangal dosha of both with a pair status (none/both/boy_only/girl_only), 7th-house facts of both (7th lord, Venus, Jupiter, D9 7th, D7 5th) and 10 years of Sade Sati windows for both. No verdict beyond the Ashtakoota one. Use for 'full compatibility check', 'complete kundli milan data', 'is this match good overall'." + idsNote,
		Title:       "Milan Summary (Matching Data Bundle)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MilanSummaryInput) (*mcp.CallToolResult, any, error) {
		q := boyGirlQuery(in.AshtakootaInput)
		setIf(q, "from", in.From)
		setIf(q, "tz", in.LocalTz)
		if in.IncludeKundliFacts {
			q.Set("include", "kundli_facts")
		}
		return callPassthrough(ctx, c, "/v1/milan/summary", q)
	})
}

// MilanDashaSyncInput is the couple plus the optional aligned timeline.
type MilanDashaSyncInput struct {
	AshtakootaInput
	Timeline        bool   `json:"timeline,omitempty" jsonschema:"optional; true adds both partners' Mahadasha/Antardasha periods on one aligned timeline"`
	MarriageWindows bool   `json:"marriage_windows,omitempty" jsonschema:"optional; true tags each timeline span favourable, neutral or avoid for marriage, with reason ids (implies timeline)"`
	From            string `json:"from,omitempty" jsonschema:"optional timeline start YYYY-MM-DD; defaults to today"`
	Years           int    `json:"years,omitempty" jsonschema:"optional timeline span in years (default 10, max 30)"`
	LocalTz         string `json:"local_tz,omitempty" jsonschema:"optional IANA zone for the timeline *_local times; defaults to the groom's birth tz"`
}

func (in MilanDashaSyncInput) toQuery() url.Values {
	q := boyGirlQuery(in.AshtakootaInput)
	setTrueIf(q, "timeline", in.Timeline)
	setTrueIf(q, "marriage_windows", in.MarriageWindows)
	setIf(q, "from", in.From)
	setIntIf(q, "years", in.Years)
	setIf(q, "tz", in.LocalTz)
	return q
}
