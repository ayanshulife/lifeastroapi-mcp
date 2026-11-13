package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Response language.
//
// The API answers every Vedic endpoint in the request language: names
// (planets, signs, nakshatras, tithis, festivals), labels, notes and the
// interpretation text. It reads the language from ?locale= and accepts
// ?lang= as an alias.
//
// Two kinds of tool expose it:
//
//   - Tools whose input struct has its own `lang` or `locale` field (the
//     horoscope, numerology, report, narrative, remedy and festival tools).
//     Their struct tag carries the description below, word for word.
//   - Data tools registered with addLocalizedTool, which adds an optional
//     `locale` input on top of the tool's own input struct and forwards it.
//
// An assistant chooses the language from the input description alone, so
// the list has to be the same everywhere. TestLanguageInputs checks that.
const (
	// vedicLanguageDesc describes the language input of every Vedic tool.
	vedicLanguageDesc = "response language: en (default), hi, mr, te, ta, kn, ml, gu (bn, pa: names only)"
	// westernLanguageDesc describes the language input of the Western
	// narrative tools, the only Western routes that are not English-only.
	westernLanguageDesc = "response language: en (default), hi, mr"
	// localizedNote is appended to the description of every tool
	// registered with addLocalizedTool.
	localizedNote = " Names and labels come back in the `locale` language."
)

type localeKey struct{}

// localeFromContext returns the `locale` argument of the running tool call,
// or "" when the tool has none or the assistant did not set it.
func localeFromContext(ctx context.Context) string {
	s, _ := ctx.Value(localeKey{}).(string)
	return s
}

// withLocale returns q with the call's locale added, unless the handler
// already set a language itself. q is not modified.
func withLocale(ctx context.Context, q url.Values) url.Values {
	loc := localeFromContext(ctx)
	if loc == "" || q.Has("locale") || q.Has("lang") {
		return q
	}
	out := make(url.Values, len(q)+1)
	for k, v := range q {
		out[k] = v
	}
	out.Set("locale", loc)
	return out
}

// addLocalizedTool registers a tool exactly like mcp.AddTool and gives it
// one more optional input, `locale`. The value is sent to the API as
// ?locale= by callPassthrough, so the handler does not have to know about
// it.
//
// Use it only for a Vedic endpoint that really answers in the request
// language. Endpoints that return nothing but numbers, times and ids
// (rahu kaal, sunrise, sarvashtakavarga, transit windows, ...) are
// registered with plain mcp.AddTool: a language input that changes nothing
// would mislead the assistant.
//
// The input struct must not have a `lang` or `locale` field of its own.
func addLocalizedTool[In any](s *mcp.Server, t *mcp.Tool, h mcp.ToolHandlerFor[In, any]) {
	schema, err := jsonschema.For[In](nil)
	if err != nil {
		panic(fmt.Sprintf("tool %s: input schema: %v", t.Name, err))
	}
	for _, name := range []string{"lang", "locale"} {
		if _, ok := schema.Properties[name]; ok {
			panic(fmt.Sprintf("tool %s: input already has %q; register it with mcp.AddTool", t.Name, name))
		}
	}
	if schema.Properties == nil {
		schema.Properties = map[string]*jsonschema.Schema{}
	}
	schema.Properties["locale"] = &jsonschema.Schema{Type: "string", Description: vedicLanguageDesc}
	schema.PropertyOrder = append(schema.PropertyOrder, "locale")

	tool := *t
	tool.InputSchema = schema
	tool.Description += localizedNote
	mcp.AddTool(s, &tool, func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		var args struct {
			Locale string `json:"locale"`
		}
		if req != nil && req.Params != nil && len(req.Params.Arguments) > 0 {
			// The arguments were already validated against the schema.
			_ = json.Unmarshal(req.Params.Arguments, &args)
		}
		if loc := strings.ToLower(strings.TrimSpace(args.Locale)); loc != "" {
			ctx = context.WithValue(ctx, localeKey{}, loc)
		}
		return h(ctx, req, in)
	})
}
