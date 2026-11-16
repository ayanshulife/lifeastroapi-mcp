# lifeastro-mcp v0.13.0

KP horary, and the KP tools return the full KP set. 321 tools.

## New tool

- `prashna_kp_horary` (30 credits): the KP horary chart for a number from 1 to 249 that the
  querent gives, cast for the moment and place of the question. Returns the horary ascendant
  (the first degree of that number's sign, star and sub), the 12 Placidus cusps with star, sub
  and sub-sub lords and their significator planets, every planet with its KP lords and house,
  the houses each planet signifies, the KP 4-step chain and the ruling planets of the moment.
  Uses the KP (Krishnamurti) ayanamsa unless `ayanamsa` is set. Takes `locale` like the other
  prashna tools.

## Changed tools

- `chart_kp_sublord` also returns the KP ruling planets for the moment and place (day lord,
  and the star, sign and sub lords of the ascendant and the Moon).
- `chart_kp_house_significator` also returns each cusp's significator planets (four KP
  levels), the houses every planet signifies and the KP 4-step chain.
- Both take an optional `ayanamsa` (use `krishnamurti` for KP work). Their `planet` and `cusp`
  inputs are now optional and ignored: the API always returned every planet and all 12 cusps.
