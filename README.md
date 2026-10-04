# lifeastro-mcp

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/ayanshulife/lifeastroapi-mcp.svg)](https://pkg.go.dev/github.com/ayanshulife/lifeastroapi-mcp)
[![Release](https://img.shields.io/github/v/release/ayanshulife/lifeastroapi-mcp)](https://github.com/ayanshulife/lifeastroapi-mcp/releases)

**Model Context Protocol (MCP) server for [LifeAstroAPI](https://lifeastroapi.com)** — exposes 308 Vedic & Western astrology tools to AI assistants like Claude Desktop, Cursor, Continue.dev, and Codex CLI.

End users don't write code, copy curl commands, or read API docs. They just chat with their AI assistant naturally and the assistant invokes the right tool when an astrology question comes up:

> **User:** Mera lagna kya hai? Birth: 15 Jan 1990, 10:30 AM, Mumbai.
>
> **Claude:** *(calls `chart_ascendant`)* Aapka lagna **Capricorn (Makara)** hai, 12°34′ par. Lagna lord Saturn 10th house mein sthit hai...

## Architecture

```
┌─────────────────────────┐
│  USER LAPTOP            │
│  Claude Desktop / Cursor│
│        ↓ stdio          │
│  lifeastro-mcp (this)   │
│        ↓ HTTPS          │
└────────│────────────────┘
         ↓ public internet
  api.lifeastroapi.com
```

`lifeastro-mcp` runs as a **subprocess of the AI client on your laptop** — not on any server. It forwards tool calls as authenticated HTTPS requests to `api.lifeastroapi.com`. Every call costs the same number of credits as a direct curl; there's no MCP surcharge.

## Installation

### macOS / Linux — Homebrew (recommended)

```bash
brew tap ayanshulife/lifeastroapi-mcp https://github.com/ayanshulife/lifeastroapi-mcp
brew install lifeastro-mcp

# Verify
lifeastro-mcp 2>&1 | head -1
# → "lifeastro-mcp: LIFEASTRO_API_KEY env var is required."
```

### Any platform — pre-built binary

1. Visit the [Releases page](https://github.com/ayanshulife/lifeastroapi-mcp/releases)
2. Download the `lifeastro-mcp-<os>-<arch>` binary matching your machine
3. Make it executable and put it on your `PATH`:

   ```bash
   # macOS / Linux
   chmod +x lifeastro-mcp-darwin-arm64
   sudo mv lifeastro-mcp-darwin-arm64 /usr/local/bin/lifeastro-mcp

   # Windows: rename to lifeastro-mcp.exe and add the folder to PATH
   ```

### Go developers — `go install`

```bash
go install github.com/ayanshulife/lifeastroapi-mcp/cmd/lifeastro-mcp@latest
```

The binary lands in `$GOBIN` (or `$GOPATH/bin`).

### Build from source

```bash
git clone https://github.com/ayanshulife/lifeastroapi-mcp.git
cd lifeastroapi-mcp
make build
# Binary: ./lifeastro-mcp
```

## Get an API key

Sign up at [lifeastroapi.com](https://lifeastroapi.com) and create a key from your dashboard. The MCP server uses the same `dv_live_<hex>` keys as direct API access — there's no separate "MCP key".

The free plan (1,500 credits a month, up to 500 a day) is enough to try most tools. Every tool call costs exactly what the same API request costs — see `meta_endpoints` for the per-route credit cost.

## Configuring your AI client

The MCP server reads its credentials from environment variables that the AI client injects at subprocess spawn. Add the config block below to your client's MCP config file.

### Claude Desktop

Config file location:
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

Add (or merge with existing `mcpServers`):

```json
{
  "mcpServers": {
    "lifeastro": {
      "command": "/usr/local/bin/lifeastro-mcp",
      "env": {
        "LIFEASTRO_API_KEY": "dv_live_REPLACE_WITH_YOUR_KEY"
      }
    }
  }
}
```

Quit and reopen Claude Desktop. The tools will appear in the tool list at the bottom of the chat window.

### Cursor

Cursor reads `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "lifeastro": {
      "command": "/usr/local/bin/lifeastro-mcp",
      "env": {
        "LIFEASTRO_API_KEY": "dv_live_REPLACE_WITH_YOUR_KEY"
      }
    }
  }
}
```

Restart Cursor (`Cmd-Shift-P → Reload Window`).

### OpenAI Codex CLI

Codex CLI reads `~/.codex/config.toml`:

```toml
[mcp_servers.lifeastro]
command = "/usr/local/bin/lifeastro-mcp"
env = { LIFEASTRO_API_KEY = "dv_live_REPLACE_WITH_YOUR_KEY" }
```

Restart Codex and the tools become available in your session.

### Continue.dev / Windsurf / other MCP clients

Any MCP-aware client supports the same shape — `command` + `env` with the absolute path to the binary and the API key. Refer to your client's docs for the config-file location.

## Tools

The server exposes **308 tools** across Vedic and Western astrology. The full list with descriptions is at [lifeastroapi.com/mcp](https://lifeastroapi.com/mcp/).

| Domain | Tools | Examples |
|---|---|---|
| Panchang | 29 | `panchang_abhijit`, `panchang_advanced`, `panchang_amrit_kalam`, `panchang_basic` |
| Birth chart (Vedic) | 22 | `chart_ascendant`, `chart_ashtakavarga`, `chart_ashtakvarga_planet`, `chart_aspects` |
| Ashtakavarga | 4 | `ashtakavarga_bhinna`, `ashtakavarga_kaksha`, `ashtakavarga_sarva`, `ashtakavarga_transit_score` |
| Dasha | 14 | `dasha_char_antardashas`, `dasha_char_current`, `dasha_char_full`, `dasha_char_major` |
| Compatibility (Milan) | 14 | `ashtakoota_breakdown`, `mangal_dosha`, `match_making_score`, `milan_ashtakoota_single` |
| Muhurta | 11 | `muhurta_best_time`, `muhurta_chandra_bala`, `muhurta_graha_pravesh`, `muhurta_naamkaran` |
| Transits & Sade Sati | 8 | `sade_sati`, `transit_ashtakavarga`, `transit_double_transit`, `transit_small_panoti` |
| Calendar, festivals & eclipses | 9 | `calendar_adhik_maas`, `calendar_month`, `calendar_ritu`, `calendar_samvatsara` |
| Planet moments | 4 | `planet_combustion_window`, `planet_ingress`, `planet_retrograde_window`, `planet_speed` |
| Prashna (horary) | 5 | `prashna_answer`, `prashna_arudha`, `prashna_chart`, `prashna_lagna_lord` |
| Numerology | 9 | `numerology_advanced`, `numerology_challenges`, `numerology_conductor`, `numerology_destiny` |
| Varshaphal | 7 | `varshaphal_chart`, `varshaphal_harsha_bala`, `varshaphal_lord`, `varshaphal_mudda_dasha` |
| Lal Kitab | 3 | `lalkitab_chart`, `lalkitab_dasha`, `lalkitab_dasha_current` |
| Remedies | 10 | `remedy_by_rule_id`, `remedy_daan_by_id`, `remedy_lalkitab_for_placement`, `remedy_lalkitab_list` |
| Narrative interpretations | 37 | `narrative_career_outlook`, `narrative_dasha_phal`, `narrative_dasha_tree_phal`, `narrative_divisional` |
| Horoscopes | 3 | `horoscope_daily`, `horoscope_monthly`, `horoscope_weekly` |
| Reports | 12 | `report_dasha_analysis`, `report_horoscope_daily`, `report_horoscope_monthly`, `report_horoscope_weekly` |
| Tarot | 5 | `tarot_card`, `tarot_cards`, `tarot_daily`, `tarot_spread` |
| Geo | 4 | `geo_place`, `geo_reverse`, `geo_search`, `geo_timezone` |
| Western astrology | 92 | `western_astrocartography_local_space`, `western_astrocartography_planet_lines`, `western_compatibility_sun_sign`, `western_composite_aspects` |
| Catalogs & discovery | 6 | `meta_ayanamsas`, `meta_endpoints`, `meta_vedic_doshas`, `meta_vedic_locales` |

### How dates work

- Tools about "today" (panchang, transits, prashna, planet moments, calendar) take an optional `date`. Leave it out and the server sends today's date in the timezone you gave.
- Tools that need a birth chart **and** a moment to evaluate (transits to a chart, Sade Sati, tara bala, personalised horoscopes) take the birth data as `date`/`time`/`tz` and the moment as the optional `on_date`.
- The per-purpose muhurta tools (`muhurta_vivah`, `muhurta_graha_pravesh`, …) score **one date** per call. To search a range for the best dates, use `muhurta_best_time`.

## Troubleshooting

**The tools don't appear in Claude Desktop.**
Check the Claude Desktop logs at `~/Library/Logs/Claude/mcp*.log`. The most common errors:

- `command not found` — the path in `command` is wrong. Use `which lifeastro-mcp` to find the absolute path and paste that.
- `LIFEASTRO_API_KEY env var is required` — the `env` block is missing or your API key has a typo.

**A tool call returns `API 401: Unauthorized`.**
Your API key is invalid, revoked, or expired. Generate a new one in the dashboard.

**A tool call returns `API 429`.**
You've hit your plan's rate limit or daily cap. Either upgrade or wait for the next reset window (free tier resets at midnight UTC).

**A tool call returns `context deadline exceeded`.**
The upstream API took longer than 60s — usually a network blip. Retry once. If it persists, check [api.lifeastroapi.com/healthz](https://api.lifeastroapi.com/healthz).

## Privacy

`lifeastro-mcp` runs entirely on your local machine. It only sends data to `api.lifeastroapi.com` (HTTPS) when an AI tool calls one of the registered tools. It does not phone home, collect telemetry, or log requests anywhere. The binary is statically linked Go — you can audit the source in this repository.

## Development

```bash
# Run tests
make test

# Format + vet
make fmt vet

# Build for current platform
make build

# Cross-compile for all release platforms
make build-all VERSION=0.11.0

# Call every tool against the real API (needs a key; spends credits).
# Reports cost 500-8000 credits each, so they are skipped unless asked for.
LIFEASTRO_API_KEY=dv_live_... make live
LIFEASTRO_API_KEY=dv_live_... make live LIVE_SKIP=        # include reports
```

## Roadmap

- **v0.11 (current):** Local stdio MCP server, 308 tools, every tool verified against the live API (`make live`)
- **v1.0 (planned):** Stable tool surface; semantic versioning kicks in
- **Remote HTTP MCP (separate project):** Hosted MCP endpoint at `mcp.lifeastroapi.com` for ChatGPT and web/mobile MCP clients — tracked separately from this repo

## License

[MIT](LICENSE) — Copyright (c) 2026 Ayanshu Life. You may use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of this software.

The MIT license applies to the MCP client code in this repository only. The LifeAstroAPI service it talks to (`api.lifeastroapi.com`) is a separate commercial product governed by its own [Terms of Service](https://lifeastroapi.com/terms).

## Support

- 🐛 Bug reports & feature requests: [GitHub Issues](https://github.com/ayanshulife/lifeastroapi-mcp/issues)
- 📖 API documentation: [api.lifeastroapi.com/docs](https://api.lifeastroapi.com/docs)
- 💬 Questions: support@lifeastroapi.com
