# lifeastro-mcp v0.12.0

Eight languages on every Vedic tool, and ten new tools that return the chart as data.

## Languages

The API now answers every Vedic endpoint in the request language — names (planets, signs,
nakshatras, tithis, festivals), labels, notes and the interpretation text. This release
makes the tools ask for it.

- One language list on every Vedic tool: `en` (default), `hi`, `mr`, `te`, `ta`, `kn`, `ml`,
  `gu`; `bn` and `pa` return names and terms in the language, longer text in English.
  Earlier releases listed only "en, hi, mr" (or left out Telugu, Malayalam and Gujarati) on
  most tools, so an assistant never asked for the other languages. 82 input descriptions
  corrected.
- 113 Vedic tools gain an optional language input. `locale` on 111 data tools: panchang (19),
  chart (22), dasha (14), milan (14), muhurta (9), calendar (5), varshaphal (5), prashna (5),
  transits (5), planet moments (4), eclipses (3), ashtakavarga (3), Lal Kitab (3). `lang` on
  `narrative_yearly_bhavishyafal` and `narrative_horoscope_daily_tamil`.
  188 of the 320 tools now take a language. Existing input names are unchanged.
- Numerology answers in the request language (it was English-only): `numerology_*` and
  `report_numerology` with `lang`.
- 25 Vedic tools have no language input on purpose — their answer is only numbers, times and
  ids: `panchang_rahu_kaal`, `panchang_yamaganda`, `panchang_gulika`, `panchang_abhijit`,
  `panchang_brahma_muhurat`, `panchang_durmuhurta`, `panchang_pradosh_kaal`,
  `panchang_sunrise_sunset`, `panchang_moonrise_moonset`, `ashtakavarga_sarva`,
  `chart_bhavabala`, `chart_yogas`, `chart_doshas`, `chart_sensitivity`, `match_making_score`,
  `milan_overlay`, `muhurta_chandra_bala`, `muhurta_tara_bala`, `sade_sati`,
  `transit_sade_sati_windows`, `transit_small_panoti`, `transit_small_panoti_history`,
  `transit_windows`, `varshaphal_muntha`, `varshaphal_solar_return_jd`.
- Ids stay in English in every language (`sun` … `ketu`, rule ids, `sign_index`, classical
  source citations).
- Western astrology, Tarot, geo and catalog tools are English. The seven `western_narrative_*`
  tools answer in `en`, `hi`, `mr`; `western_narrative_firdaria` used to list English only.
- Without a language the API default applies: English, or the language set on the API key.
- `narrative_horoscope_daily_tamil` — the text is English unless `lang` is set; pass
  `lang: "ta"` for Tamil text.

Example — "આજનું પંચાંગ, મુંબઈ" → `panchang_today` with `locale: "gu"` returns the tithi,
nakshatra, yoga, karana and weekday names in Gujarati.

## New tools

- `chart_kundli_facts` — the whole birth chart as structured data in one call: lagna, every
  graha (sign, degree, house, nakshatra, dignity, combustion, shadbala, divisional
  placements), every house, chara karakas, avakhada.
- `chart_yogas`, `chart_doshas` — yogas and doshas as data: id, grahas and houses that form
  it, strength, cancellation reasons, classical source.
- `chart_sensitivity` — how many minutes the birth time can move before the lagna, a varga
  lagna or the Moon's nakshatra pada changes.
- `transit_sade_sati_windows` — dated Sade Sati and Dhaiya windows for up to 60 years, with
  phases and retrograde re-entries.
- `transit_windows` — every stay of one planet in a sign over a range of years, with the house
  from the natal Moon and lagna when a birth chart is given.
- `remedy_for_graha` — remedies for one graha by reason (debilitated, combust, dusthana,
  weak shadbala, dasha, or a dosha).
- `milan_overlay` — one partner's grahas in the other's houses, shared signs and nakshatras,
  Moon sign relation, lagna lords.
- `milan_summary` — all kundli matching data for a couple in one call.
- `narrative_prashna` — a written Prashna (horary) reading for a topic (career, financial,
  health, litigation, marriage, missing_person, property, travel) or a question house.

Total tools: 320.

## Changed tools

- `milan_dasha_sync` — `timeline: true` adds both partners' Mahadasha/Antardasha on one
  aligned timeline (`from`, `years`); `marriage_windows: true` tags each span favourable,
  neutral or avoid for marriage, with reason ids.
- `narrative_milan` — fixed: the groom and the bride were sent to the API the wrong way round.
- `narrative_dosha_by_id` — the example ids in the input description were not valid ids and
  returned an empty reading; they now show `mangal`, `kaalsarp`, `pitra`.
- `narrative_dasha_phal`, `narrative_transit_phal`, `varshaphal_lord`, `chart_nakshatra` —
  descriptions corrected to what the tools return.
- Two additive response fields from the API pass through: `factors[].id` on the muhurta tools
  and `avoidance_text` on `panchang_panchaka`.

## Other changes

- Docker image `ghcr.io/ayanshulife/lifeastroapi-mcp` (linux/amd64 and linux/arm64), built on
  every tag: `docker run -i --rm -e LIFEASTRO_API_KEY=... ghcr.io/ayanshulife/lifeastroapi-mcp:0.12.0`.
- `server.json` for the official MCP Registry and `glama.json`.
- `make live` takes `LIFEASTRO_LIVE_LOCALE=gu` (or another language): it passes the language
  to every tool that takes one and lists the tools that answered without text in it.

## Requires

LifeAstroAPI backend of 2026-11-13 or later (request-language responses on every Vedic
endpoint, Gujarati and Malayalam interpretation text). With an older backend the new `locale`
inputs are accepted and the answer stays in English.

## Release checklist

1. Push the tag `v0.12.0` (GitHub Desktop: History → right-click the latest commit →
   Create Tag… → `v0.12.0` → Push origin). The Release workflow builds all five
   binaries in CI and creates a **draft** release with them and a checksums file; the Docker
   workflow pushes `ghcr.io/ayanshulife/lifeastroapi-mcp:0.12.0` and `:latest`.
2. Open the draft release, paste this file as the description, and publish it.
3. Update `HomebrewFormula/lifeastro-mcp.rb`: `version "0.12.0"` and the sha256 values from
   the **release's** `lifeastro-mcp-checksums.txt` (not the local `dist/` one: CI and local
   builds can differ byte for byte). Commit and push.
4. Publish `server.json` (already at 0.12.0) to the official MCP Registry once the image for
   0.12.0 is on GHCR.
5. Website: set `MCP_VERSION` to `0.12.0` on the MCP page (tool count stays 320).
