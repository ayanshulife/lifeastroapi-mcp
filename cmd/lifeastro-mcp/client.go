package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiClient is a thin wrapper around http.Client that knows how to
// talk to api.lifeastroapi.com — base URL, bearer auth, JSON decode,
// useful error messages.
//
// Connection-pooled by design: multiple goroutines from the MCP server
// can call concurrent tools without thrashing TCP. The default
// http.Client transport already pools, but we tune the timeout to a
// reasonable per-request limit (15s — most endpoints respond in
// under 100ms; PDF endpoints take longer but they're not exposed in
// MCP v1).
type apiClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// maxResponseBytes caps how much of an upstream response is read.
const maxResponseBytes = 32 << 20

// apiErrorMessage extracts "code: message (request_id ...)" from an API
// error body, or "" when the body is not a recognised error shape.
func apiErrorMessage(body []byte) string {
	var nested struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &nested); err == nil && nested.Error.Message != "" {
		msg := nested.Error.Message
		if nested.Error.Code != "" {
			msg = nested.Error.Code + ": " + msg
		}
		if nested.Error.RequestID != "" {
			msg += " (request_id " + nested.Error.RequestID + ")"
		}
		return msg
	}
	var flat struct {
		Error   string `json:"error"`
		Code    string `json:"code"`
		Details string `json:"details"`
	}
	if err := json.Unmarshal(body, &flat); err == nil && flat.Error != "" {
		msg := flat.Error
		if flat.Code != "" {
			msg = flat.Code + ": " + msg
		}
		if flat.Details != "" {
			msg += " — " + flat.Details
		}
		return msg
	}
	return ""
}

func newAPIClient(baseURL, apiKey, version string) *apiClient {
	return &apiClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			// Composed reports (kundli/brihad, dasha-analysis) take
			// several seconds server-side; 15s cut them off.
			Timeout: 60 * time.Second,
		},
		userAgent: "lifeastro-mcp/" + version,
	}
}

// get performs an authenticated GET against the upstream API and
// decodes the JSON response into out. Query parameters are passed as
// a url.Values map so callers don't have to construct query strings
// manually.
//
// On 4xx/5xx, the upstream JSON error body (if any) is surfaced in
// the returned error so the AI client sees a useful message instead
// of "request failed". This is critical for debugging — when an AI
// assistant says "the tool failed", users want to know why.
func (c *apiClient) get(ctx context.Context, path string, params url.Values, out any) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", path, err)
	}
	defer resp.Body.Close()

	// 32 MiB cap. The old 1 MiB cap silently truncated the larger
	// responses (kundli/brihad report, 30-day transit calendar), which
	// then failed JSON decoding with "unexpected end of JSON input".
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if len(body) > maxResponseBytes {
		return fmt.Errorf("response from %s exceeds %d MiB", path, maxResponseBytes>>20)
	}

	if resp.StatusCode >= 400 {
		// Surface the API's structured error so the assistant sees why
		// a call failed. The API returns
		//   {"error": {"code": "...", "message": "...", "request_id": "..."}}
		// (an older flat {"error": "...", "code": "..."} shape is still
		// accepted). Unknown shapes fall through to the raw body.
		if msg := apiErrorMessage(body); msg != "" {
			return fmt.Errorf("API %d %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("API %d: %s", resp.StatusCode, string(body))
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
