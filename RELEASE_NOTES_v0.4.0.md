# lifeastro-mcp v0.4.0

Festival and calendar tools now match printed Hindu calendars.

## Changed tools

- `festivals_month`, `festivals_on_date` — now location-aware: optional `lat`, `lon`, `tz`
  (IANA). Dates follow the classical udaya / pradosh / nishita / moonrise / aparahna rules
  and reproduce the Drik Panchang 2027 calendars of 14 cities (Delhi, Chennai, Bengaluru,
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

## Requires

LifeAstroAPI backend with the 2026-10 festival engine (`/v1/festivals/*` location params,
`/v1/panchang/monthly?include=sheet`).

## Release checklist

1. `git tag v0.4.0 && git push origin v0.4.0`
2. Attach `dist/lifeastro-mcp-*` and `dist/lifeastro-mcp-checksums.txt` to the GitHub release.
3. Update `HomebrewFormula/lifeastro-mcp.rb`: `version "0.4.0"` + the five sha256 values
   from `dist/lifeastro-mcp-checksums.txt`; commit and push.
