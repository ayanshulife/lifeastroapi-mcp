# lifeastro-mcp v1.1.0

Ashtottari dasha, Jaimini chart data and six more muhurta tools. 332 tools.
Every change is additive: with the same configuration the server lists the same tools as 1.0.0
plus the eleven new ones, and every existing input still works.

## New tools

Dasha:

- `dasha_ashtottari_full`: the 108-year Ashtottari cycle of 8 lords with start and end dates,
  the balance at birth, and the two classical applicability conditions (Rahu from the lagna
  lord; day or night birth in krishna or shukla paksha), reported as information only.
  `depth` 2 adds antardashas, 3 pratyantardashas. `scheme` 28 (default, with Abhijit) or 27;
  `antar_start` `maha_lord` (default) or `next_lord`.
- `dasha_ashtottari_current`: the Ashtottari mahadasha and antardasha running now, or at `at`.
  Same `scheme` and `antar_start` options.

Birth chart (Jaimini):

- `chart_arudha_padas`: A1 to A12 with the working for each house; A1 is the Arudha Lagna, A12
  the Upapada. Options `exception` (`parashari` default, or `none`), `scorpio_lord` (`mars` or
  `ketu`), `aquarius_lord` (`saturn` or `rahu`).
- `chart_upapada`: the Upapada Lagna with its lord, the grahas in it and the grahas aspecting it
  by graha drishti and rashi drishti. Same options as `chart_arudha_padas`.
- `chart_argala`: argala and virodha argala on every house, or on one with `house`, with a
  status for each pair and vipareeta argala. Options `include_eighth`, `node_reversal`
  (`ketu` default, `rahu_ketu`, `none`) and `vipareeta_min_malefics` (default 3).

Muhurta, one date per call like the other muhurta tools:

- `muhurta_vahan`: buying a vehicle or using it for the first time.
- `muhurta_bhoomi_pujan`: ground-breaking or laying a foundation; also returns the corner the
  Rahu head faces in that solar month and the corner where digging starts (information only).
- `muhurta_upanayan`: the sacred thread ceremony.
- `muhurta_annaprashan`: a child's first solid food; optional `birth_date` and `sex` add an age
  note.
- `muhurta_vidyarambh`: a child's first letters; optional `birth_date` adds an age note.
- `muhurta_karnavedha`: ear piercing; optional `birth_date` adds an age note.

The four sanskar tools (upanayan, annaprashan, vidyarambh, karnavedha) also return `notes`:
informational sentences such as the traditional age and festival days. They never change the
score.

All eleven take `locale` like the other dasha, chart and muhurta tools.

## Classical prohibitions in muhurta results

- Muhurta results now include `prohibited_by`: the classical prohibitions of that rite that apply
  to the moment (Guru or Shukra asta, adhika masa, Bhadra, Hari Shayana, Dakshinayana). When it
  is not empty the grade is `inauspicious` and `is_acceptable` is false, whatever the score.
- The rites that have prohibitions: vivah (Guru or Shukra asta, adhika masa, Bhadra), graha
  pravesh (Guru or Shukra asta, adhika masa), bhoomi pujan (Bhadra, Guru or Shukra asta, adhika
  masa), vahan (Bhadra), upanayan (Guru or Shukra asta, adhika masa, Dakshinayana, Bhadra),
  karnavedha (Guru or Shukra asta, adhika masa, Hari Shayana, Bhadra), vidyarambh (adhika masa,
  Bhadra), annaprashan (Bhadra). For naamkaran, vyapar and yatra it is always empty.

## Changed tools

- `muhurta_best_time` takes `purpose`: `vivah`, `graha_pravesh`, `naamkaran`, `vyapar`, `yatra`,
  `vahan`, `bhoomi_pujan` or `general` (the default, as before). Each purpose is scored with the
  rules of its own muhurta tool, and a moment under one of its prohibitions is never returned,
  so fewer windows, or none, can come back. It also takes `direction` (used with
  `purpose=yatra`) and `birth_date` (the child's birth date with `purpose=naamkaran`).
- `muhurta_yatra` takes an optional `direction` (N, NE, E, SE, S, SW, W, NW; full names such as
  north also work). With it the weekday's disha shool is checked; without it there is no
  direction check (until now a call without a direction was checked for east).
- The descriptions of `muhurta_vivah`, `muhurta_graha_pravesh`, `muhurta_naamkaran` and
  `muhurta_vyapar` now point to `muhurta_best_time` with the matching purpose.
