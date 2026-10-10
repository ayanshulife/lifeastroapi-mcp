# lifeastro-mcp v0.11.0

Festival and calendar tools now match printed Hindu calendars.

## Changed tools

- `festivals_month`, `festivals_on_date` — now location-aware: optional `lat`, `lon`, `tz`
  (IANA). Dates follow the classical udaya / pradosh / nishita / moonrise / aparahna rules
  and reproduce the printed panchang calendars of 14 cities (Delhi, Chennai, Bengaluru,
  Hyderabad, Pune, Dubai, Singapore, London, Melbourne, Sydney, San Francisco, Seattle,
  Toronto, Brampton) entry for entry. Every named Ekadashi (smarta + Vaishnava/Gauna with
  parana window), Pradosh by weekday, Sankashti/Vinayaka Chaturthi, every Purnima and
  Amavasya, Chandra Darshana, all Sankrantis, regional festivals (Tamil, Kerala, Marathi,
  Telugu, Bengali, Gujarati, Punjabi), jayantis, national days and eclipses.
  New `locale` (hi, mr, ta, kn, bn, gu, pa) and `month_system` (purnimanta | amanta).
  The old `region` input (never supported by the API) is removed.
- `panchang_monthly` — `tz` now accepts IANA names (DST-correct); new `alt` (elevation),
  `sheet: true` returns the complete printable calendar page (tithi / nakshatra / moon-sign
  end-times in "27:07+" notation, kshaya days, moonrise/moonset, Hindu month, Vikram/Shaka
  samvat, moon phase, festivals per day, page header), plus `locale` and `month_system`.

- New `festivals_vrat` — year-wise Vrat & Upavas lists (Ekadashi, Pradosh, Sankashti, Purnima, Amavasya, Shraddha, Janmashtami, …).
- New `calendar_solstices` — equinoxes and solstices with Uttarayana/Dakshinayana and day length.
- `eclipses_solar`, `eclipses_lunar`, `eclipses_all` — with lat/lon (and `tz`) each eclipse carries a
  `local` block: sparsha/madhya/moksha contact times in local time, visibility, and sutak kaal.

Total tools: 310.

## Other changes

- The server no longer exits when `LIFEASTRO_API_KEY` is missing. It starts, lists all
  310 tools (so MCP registries and inspectors can read them), and each tool call returns
  a clear "set LIFEASTRO_API_KEY" message until a key is configured.
- `smithery.yaml` and a `Dockerfile` for container-based MCP registries.

## Requires

LifeAstroAPI backend with the 2026-10 festival engine (`/v1/festivals/*` location params,
`/v1/panchang/monthly?include=sheet`).

## Release checklist

1. Push the tag `v0.11.0` (GitHub Desktop: History → right-click the latest commit →
   Create Tag… → `v0.11.0` → Push origin). The Release workflow builds all five
   binaries in CI and creates a **draft** release with them and a checksums file.
2. Open the draft release, paste this file as the description, and publish it.
3. Update `HomebrewFormula/lifeastro-mcp.rb`: `version "0.11.0"` and the sha256 values from
   the **release's** `lifeastro-mcp-checksums.txt` (not the local `dist/` one: CI and local
   builds can differ byte for byte). Commit and push.
