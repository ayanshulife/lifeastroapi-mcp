package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerV11Tools closes the gap found by the 2026-10 API↔MCP audit:
// routes that exist in the API catalog but had no tool.
//
//	Chart (1):     /v1/chart/ashtakvarga/{planet}
//	Narrative (1): /v1/vedic/narrative/reference/{family}
//	Catalogs (6):  /v1/meta/vedic/yogas, /doshas, /locales, /narrative-axes,
//	               /v1/meta/ayanamsas, /v1/meta/endpoints
//
// The yoga and dosha catalogs matter beyond coverage: narrative_yoga_by_id
// and narrative_dosha_by_id have always told the assistant to look IDs up
// with meta_vedic_yogas / meta_vedic_doshas, tools that did not exist.
//
// Deliberately not exposed: /v1/vedic/narrative/reference without a family
// (~800 KB, far beyond what an assistant can use in one result), /v1/docs
// (the OpenAPI document) and /v1/meta/version (operational).
//
// Total new: 8 tools.
func registerV11Tools(s *mcp.Server, c *apiClient) {
	registerChartAshtakvargaPlanet(s, c)
	registerNarrativeReference(s, c)
	registerMetaVedicYogas(s, c)
	registerMetaVedicDoshas(s, c)
	registerMetaVedicLocales(s, c)
	registerMetaVedicNarrativeAxes(s, c)
	registerMetaAyanamsas(s, c)
	registerMetaEndpoints(s, c)
}

type ChartAshtakvargaPlanetInput struct {
	BirthInput
	Planet string `json:"planet" jsonschema:"planet name: Sun, Moon, Mars, Mercury, Jupiter, Venus or Saturn (Rahu and Ketu have no ashtakavarga)"`
}

func registerChartAshtakvargaPlanet(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "chart_ashtakvarga_planet",
		Description: "Get one planet's Bhinnashtakavarga row from the chart endpoint — the benefic points (bindus) that planet holds in each of the 12 signs, with its total. A compact single-planet alternative to the full ashtakavarga table. Use for 'Jupiter's ashtakavarga points', 'how many bindus does Saturn have in each sign'.",
		Title:       "Ashtakavarga for One Planet",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartAshtakvargaPlanetInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/chart/ashtakvarga/"+url.PathEscape(strings.TrimSpace(in.Planet)), in.BirthInput.toQuery())
	})
}

type NarrativeReferenceInput struct {
	Family string `json:"family" jsonschema:"glossary family to read, e.g. yoga, dosha, argala — the short name, or a full rule prefix such as vedic.L1.yoga"`
	Locale string `json:"locale,omitempty" jsonschema:"response language: en (default), hi, or another locale from meta_vedic_locales"`
}

func registerNarrativeReference(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "narrative_reference",
		Description: "Read the authored Vedic astrology glossary for one concept family — explanatory text that defines a concept (what a yoga or dosha is, how a technique works) rather than interpreting one person's chart. No birth data needed. Use for 'explain what Gaja Kesari yoga means', 'what is argala', 'definitions of the doshas'. A family can be large; prefer narrative_yoga_by_id or narrative_dosha_by_id when you only need one entry.",
		Title:       "Vedic Glossary by Family",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in NarrativeReferenceInput) (*mcp.CallToolResult, any, error) {
		q := url.Values{}
		if in.Locale != "" {
			q.Set("locale", in.Locale)
		}
		return callPassthrough(ctx, c, "/v1/vedic/narrative/reference/"+url.PathEscape(strings.TrimSpace(in.Family)), q)
	})
}

// NoInput is the input of tools that take no arguments.
type NoInput struct{}

func registerMetaVedicYogas(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "meta_vedic_yogas",
		Description: "List every yoga the API can detect and narrate, with its ID, a one-line definition, the life domains it touches and the planets involved. Free (0 credits). Call this to find the ID that narrative_yoga_by_id needs, or to answer 'which yogas do you support'.",
		Title:       "Yoga Catalog",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/meta/vedic/yogas", nil)
	})
}

func registerMetaVedicDoshas(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "meta_vedic_doshas",
		Description: "List every dosha the API can detect and narrate (Mangal, Kaal Sarp and its sub-types, Pitru and others), with its ID, the life domains it touches and the planets involved. Free (0 credits). Call this to find the ID that narrative_dosha_by_id needs.",
		Title:       "Dosha Catalog",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/meta/vedic/doshas", nil)
	})
}

func registerMetaVedicLocales(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "meta_vedic_locales",
		Description: "List the languages the narrative endpoints can answer in, with how much of the interpretation corpus each one covers. Free (0 credits). Use before passing a lang/locale to a narrative tool, or to answer 'which languages are supported'.",
		Title:       "Narrative Languages",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/meta/vedic/locales", nil)
	})
}

func registerMetaVedicNarrativeAxes(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "meta_vedic_narrative_axes",
		Description: "List the axes along which narrative text can be varied (such as locale and regional pack). Free (0 credits). Rarely needed directly — use it to discover which variation parameters the narrative tools understand.",
		Title:       "Narrative Variation Axes",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/meta/vedic/narrative-axes", nil)
	})
}

type MetaAyanamsasInput struct {
	Date string `json:"date,omitempty" jsonschema:"optional date to evaluate the ayanamsa values on, YYYY-MM-DD; defaults to today"`
	Tz   string `json:"tz,omitempty" jsonschema:"optional IANA timezone for the date; defaults to UTC"`
}

func registerMetaAyanamsas(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "meta_ayanamsas",
		Description: "List the ayanamsa systems the API supports (Lahiri, Raman, Krishnamurti/KP, Fagan-Bradley and others) with each one's value in degrees on a date, and which is the default. Free (0 credits). Use for 'what is the Lahiri ayanamsa today', 'difference between Lahiri and KP ayanamsa'.",
		Title:       "Supported Ayanamsas",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MetaAyanamsasInput) (*mcp.CallToolResult, any, error) {
		q := url.Values{}
		tz := tzOrUTC(in.Tz)
		q.Set("tz", tz)
		q.Set("date", dateOrToday(in.Date, tz))
		return callPassthrough(ctx, c, "/v1/meta/ayanamsas", q)
	})
}

func registerMetaEndpoints(s *mcp.Server, c *apiClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "meta_endpoints",
		Description: "List every API route with its credit cost and group. Free (0 credits). Use to answer 'how many credits does X cost', 'what will this request cost me', or to estimate the cost of a plan before running several tools.",
		Title:       "API Catalog & Credit Costs",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ NoInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/meta/endpoints", nil)
	})
}

// westernReceptionPlanets are the seven classical planets the reception
// endpoint scores; it takes their tropical longitudes as query params.
var westernReceptionPlanets = []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}

// westernReceptions computes mutual receptions from birth data. The API's
// reception endpoint takes planet longitudes, not a birth moment, so this
// first reads the natal planets and then feeds their longitudes in. Two
// upstream requests, so it costs both endpoints' credits.
func westernReceptions(ctx context.Context, c *apiClient, in WesternBirthInput) (*mcp.CallToolResult, any, error) {
	var natal struct {
		Planets []struct {
			Name      string  `json:"name"`
			Longitude float64 `json:"longitude"`
		} `json:"planets"`
	}
	var raw json.RawMessage
	if err := c.get(ctx, "/v1/western/natal/planets", in.toQuery(), &raw); err != nil {
		return errResult(err)
	}
	if err := json.Unmarshal(raw, &natal); err != nil {
		return errResult(fmt.Errorf("decode natal planets: %w", err))
	}
	q := url.Values{}
	for _, want := range westernReceptionPlanets {
		for _, p := range natal.Planets {
			if strings.EqualFold(p.Name, want) {
				q.Set(strings.ToLower(want), fmt.Sprintf("%.6f", p.Longitude))
			}
		}
	}
	if len(q) < 2 {
		return errResult(fmt.Errorf("natal planets response held fewer than two classical planets"))
	}
	return callPassthrough(ctx, c, "/v1/western/dignities/receptions", q)
}
