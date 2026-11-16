package main

import (
	"context"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerV13Tools adds KP horary (2026-11-15).
//
//	Prashna (1): /v1/prashna/kp-horary
//
// chart_kp_sublord and chart_kp_house_significator (tools_v06.go) return the
// new KP fields (ruling planets; significators and the 4-step chain), and
// their planet / cusp inputs became optional: the API never read them.
//
// Total new: 1 tool.
func registerV13Tools(s *mcp.Server, c *apiClient) {
	registerPrashnaKPHorary(s, c)
}

// KPHoraryInput is the moment and place of judgement plus the querent's number.
type KPHoraryInput struct {
	PrashnaInput
	Number   int    `json:"number" jsonschema:"the KP horary number the querent gives, a whole number from 1 to 249"`
	Ayanamsa string `json:"ayanamsa,omitempty" jsonschema:"optional ayanamsa; defaults to krishnamurti, the KP ayanamsa"`
}

func registerPrashnaKPHorary(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "prashna_kp_horary",
		Description: "Cast a KP (Krishnamurti Paddhati) horary chart for a number from 1 to 249 that the querent gives (30 credits), for the moment and place of the question: the horary ascendant (the first degree of that number's sign, star and sub), the 12 Placidus cusps with star, sub and sub-sub lords and their significator planets, every planet with its KP lords and house, the houses each planet signifies, the KP 4-step chain and the ruling planets of the moment. Uses the KP (Krishnamurti) ayanamsa unless another is asked for. Use for 'KP prashna with number', 'horary number 1 to 249', 'KP horary chart', 'krishnamurti prashna'.",
		Title:       "KP Horary (1-249)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in KPHoraryInput) (*mcp.CallToolResult, any, error) {
		q := in.PrashnaInput.toQuery()
		q.Set("number", strconv.Itoa(in.Number))
		setIf(q, "ayanamsa", in.Ayanamsa)
		return callPassthrough(ctx, c, "/v1/prashna/kp-horary", q)
	})
}
