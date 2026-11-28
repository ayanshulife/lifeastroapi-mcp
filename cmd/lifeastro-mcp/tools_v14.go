package main

import (
	"context"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerV14Tools adds Ashtottari dasha, the Jaimini chart data and six
// muhurta validators (2026-11-28). Every route reads the birth chart or the
// event moment unprefixed (lat, lon, date, time, tz), like their siblings.
//
//	Dasha (2):   /v1/dasha/ashtottari, /v1/dasha/ashtottari/current
//	Chart (3):   /v1/chart/arudha-padas, /v1/chart/upapada, /v1/chart/argala
//	Muhurta (6): /v1/muhurta/vahan, /bhoomi-pujan, /upanayan, /annaprashan,
//	             /vidyarambh, /karnavedha
//
// Changed in the same release: muhurta_best_time takes purpose (four new
// ones: vyapar, yatra, vahan, bhoomi_pujan), direction and birth_date;
// muhurta_yatra takes an optional direction with no default (tools_v02.go,
// tools_v06.go).
//
// Total new: 11 tools.
func registerV14Tools(s *mcp.Server, c *apiClient) {
	registerDashaAshtottariFull(s, c)
	registerDashaAshtottariCurrent(s, c)
	registerChartArudhaPadas(s, c)
	registerChartUpapada(s, c)
	registerChartArgala(s, c)
	registerMuhurtaVahan(s, c)
	registerMuhurtaBhoomiPujan(s, c)
	registerMuhurtaUpanayan(s, c)
	registerMuhurtaAnnaprashan(s, c)
	registerMuhurtaVidyarambh(s, c)
	registerMuhurtaKarnavedha(s, c)
}

// --------------------------------------------------------------------
// Dasha: Ashtottari
// --------------------------------------------------------------------

type DashaAshtottariInput struct {
	BirthInput
	Depth      int    `json:"depth,omitempty" jsonschema:"optional levels to return: 1 = mahadashas (default), 2 = with antardashas, 3 = with pratyantardashas"`
	Scheme     int    `json:"scheme,omitempty" jsonschema:"optional count of Saturn's nakshatra group: 28 (default, the Brihat Parashara Hora Shastra count with Abhijit) or 27 (27 nakshatras, linear over Saturn's 40° arc); the two differ only when the birth Moon is between 253°20' and 293°20' sidereal"`
	AntarStart string `json:"antar_start,omitempty" jsonschema:"optional lord of the first sub-period: maha_lord (default, the parent period's lord) or next_lord (the lord after it)"`
}

type DashaAshtottariCurrentInput struct {
	DashaCurrentInput
	Scheme     int    `json:"scheme,omitempty" jsonschema:"optional count of Saturn's nakshatra group: 28 (default, the Brihat Parashara Hora Shastra count with Abhijit) or 27 (27 nakshatras, linear over Saturn's 40° arc); the two differ only when the birth Moon is between 253°20' and 293°20' sidereal"`
	AntarStart string `json:"antar_start,omitempty" jsonschema:"optional lord of the first sub-period: maha_lord (default, the parent period's lord) or next_lord (the lord after it)"`
}

func registerDashaAshtottariFull(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "dasha_ashtottari_full",
		Description: "Get the full Ashtottari Dasha sequence for a birth chart — the 108-year cycle of 8 lords (Sun 6, Moon 15, Mars 8, Mercury 17, Saturn 10, Jupiter 19, Rahu 12, Venus 21 years; Ketu has none) with start/end dates, the balance at birth, and the two classical applicability conditions (Rahu from the lagna lord; day or night birth in krishna or shukla paksha), reported as information only. depth=2 adds antardashas, depth=3 pratyantardashas. Use for 'ashtottari dasha', 'my 108-year dasha', 'is ashtottari applicable to my chart'.",
		Title:       "Ashtottari Dasha Full Timeline",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in DashaAshtottariInput) (*mcp.CallToolResult, any, error) {
		q := in.BirthInput.toQuery()
		setIntIf(q, "depth", in.Depth)
		setIntIf(q, "scheme", in.Scheme)
		setIf(q, "antar_start", lowerTrim(in.AntarStart))
		return callPassthrough(ctx, c, "/v1/dasha/ashtottari", q)
	})
}

func registerDashaAshtottariCurrent(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "dasha_ashtottari_current",
		Description: "Get the currently-running Ashtottari Mahadasha and Antardasha (108-year, 8-lord system) for a birth chart, or the ones running at a given moment, with start/end dates. A moment outside the chart's 108-year cycle is an error; the cycle is not repeated. Use for 'ashtottari dasha now', 'which ashtottari period am I in'.",
		Title:       "Current Ashtottari Dasha",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in DashaAshtottariCurrentInput) (*mcp.CallToolResult, any, error) {
		q := in.BirthInput.toQuery()
		setIf(q, "at", in.At)
		setIntIf(q, "scheme", in.Scheme)
		setIf(q, "antar_start", lowerTrim(in.AntarStart))
		return callPassthrough(ctx, c, "/v1/dasha/ashtottari/current", q)
	})
}

// --------------------------------------------------------------------
// Chart: Jaimini arudha padas, Upapada, argala
// --------------------------------------------------------------------

// ChartArudhaInput is the birth chart plus the pada rule choices shared by
// arudha-padas and upapada.
type ChartArudhaInput struct {
	BirthInput
	Exception    string `json:"exception,omitempty" jsonschema:"optional rule when the counted pada falls in the bhava itself or the 7th from it: parashari (default, take the 10th from it) or none (the counted sign stands)"`
	ScorpioLord  string `json:"scorpio_lord,omitempty" jsonschema:"optional lord taken for Scorpio: mars (default) or ketu"`
	AquariusLord string `json:"aquarius_lord,omitempty" jsonschema:"optional lord taken for Aquarius: saturn (default) or rahu"`
}

func (in ChartArudhaInput) toQuery() url.Values {
	q := in.BirthInput.toQuery()
	setIf(q, "exception", lowerTrim(in.Exception))
	setIf(q, "scorpio_lord", lowerTrim(in.ScorpioLord))
	setIf(q, "aquarius_lord", lowerTrim(in.AquariusLord))
	return q
}

func registerChartArudhaPadas(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "chart_arudha_padas",
		Description: "Get the 12 Jaimini arudha padas (bhava padas A1-A12) of a birth chart with the working for each house: bhava sign, its lord and the lord's sign, the count, whether the exception applied, and the arudha sign and its house from the lagna. A1 is the Arudha Lagna (AL), A12 the Upapada (UL). Use for 'what is my arudha lagna', 'arudha padas in my chart', 'A7 / darapada'.",
		Title:       "Arudha Padas (A1-A12)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartArudhaInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/chart/arudha-padas", in.toQuery())
	})
}

func registerChartUpapada(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "chart_upapada",
		Description: "Get the Upapada Lagna (UL, the arudha of the 12th house) of a birth chart: its sign and house from the lagna, the lord of that sign with its sign, house and place counted from the UL, the grahas in the UL sign, and the grahas aspecting it by graha drishti and by rashi drishti. States facts only, no judgement. Use for 'what is my upapada lagna', 'upapada for marriage', 'UL in my chart'.",
		Title:       "Upapada Lagna",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartArudhaInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/chart/upapada", in.toQuery())
	})
}

type ChartArgalaInput struct {
	BirthInput
	House                int    `json:"house,omitempty" jsonschema:"optional house 1 to 12 from the lagna to return alone; leave out for all twelve houses"`
	IncludeEighth        bool   `json:"include_eighth,omitempty" jsonschema:"optional: also return the 8th-place secondary argala, obstructed from the 6th (default false)"`
	NodeReversal         string `json:"node_reversal,omitempty" jsonschema:"optional rule for counting the places in reverse: ketu (default, for a house that holds Ketu), rahu_ketu (a house that holds Rahu or Ketu) or none"`
	VipareetaMinMalefics int    `json:"vipareeta_min_malefics,omitempty" jsonschema:"optional number of natural malefics in the 3rd place needed for vipareeta argala, 1 to 9 (default 3)"`
}

func registerChartArgala(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "chart_argala",
		Description: "Get Jaimini argala (planetary intervention) and virodha argala (obstruction) on every house of a birth chart, or on one house: for each house the argala from the 2nd, 4th, 11th and 5th and its obstruction from the 12th, 10th, 3rd and 9th, the grahas in each place with their natural nature, and a status (none, unobstructed, prevails, obstructed, tied) with whether the argala stands; also vipareeta argala from malefics in the 3rd. Use for 'argala on my 10th house', 'which planets intervene on my 7th house', 'virodha argala'.",
		Title:       "Argala (Jaimini)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ChartArgalaInput) (*mcp.CallToolResult, any, error) {
		q := in.BirthInput.toQuery()
		setIntIf(q, "house", in.House)
		setTrueIf(q, "include_eighth", in.IncludeEighth)
		setIf(q, "node_reversal", lowerTrim(in.NodeReversal))
		setIntIf(q, "vipareeta_min_malefics", in.VipareetaMinMalefics)
		return callPassthrough(ctx, c, "/v1/chart/argala", q)
	})
}

// --------------------------------------------------------------------
// Muhurta validators
// --------------------------------------------------------------------

// prohibitedNote builds the sentence every new muhurta tool carries about
// prohibited_by, naming that rite's classical prohibitions.
func prohibitedNote(prohibitions string) string {
	return " prohibited_by lists the classical prohibitions that apply (for this rite: " + prohibitions + "); when it is not empty the grade is inauspicious and is_acceptable is false whatever the score."
}

// sanskarNotesNote is added to the sanskar tools, whose result also has notes.
const sanskarNotesNote = " notes holds informational sentences (traditional age, festival days) that never change the score."

// MuhurtaChildInput adds the child's optional birth date, used only for an
// age note.
type MuhurtaChildInput struct {
	MuhurtaDateInput
	BirthDate string `json:"birth_date,omitempty" jsonschema:"optional child's date of birth, YYYY-MM-DD; adds an age note and never changes the score"`
}

func (in MuhurtaChildInput) toQuery() url.Values {
	q := in.MuhurtaDateInput.toQuery()
	setIf(q, "birth_date", in.BirthDate)
	return q
}

// MuhurtaAnnaprashanInput adds the child's optional sex for the age note.
type MuhurtaAnnaprashanInput struct {
	MuhurtaChildInput
	Sex string `json:"sex,omitempty" jsonschema:"optional child's sex, male or female; with birth_date the note gives the child's running month and the months tradition names for a boy or a girl"`
}

func registerMuhurtaVahan(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "muhurta_vahan",
		Description: "Score how auspicious a specific date (and optional time) is for buying a vehicle or using it for the first time (vahan) at a location — returns a 0-100 score, a grade, and the factor breakdown (tithi, vara, nakshatra, Rahu Kaal, Yamaganda, Durmuhurta, Bhadra)." + prohibitedNote("Bhadra") + " Evaluates ONE date per call; to search a range use muhurta_best_time with purpose=vahan. Use for 'is this a good day to buy a car', 'vehicle purchase muhurta', 'bike delivery date check'.",
		Title:       "Vahan (Vehicle) Muhurta",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MuhurtaDateInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/muhurta/vahan", in.toQuery())
	})
}

func registerMuhurtaBhoomiPujan(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "muhurta_bhoomi_pujan",
		Description: "Score how auspicious a specific date (and optional time) is for bhoomi pujan (ground-breaking or laying a foundation) at a location — returns a 0-100 score, a grade, the factor breakdown (tithi, vara, nakshatra, solar month, paksha, Rahu Kaal, Yamaganda, Durmuhurta, Bhadra, Guru/Shukra asta, adhika masa), and an informational direction block for that solar month (the corner the Rahu head faces and the corner where digging starts)." + prohibitedNote("Bhadra, Guru or Shukra asta, adhika masa") + " Evaluates ONE date per call; to search a range use muhurta_best_time with purpose=bhoomi_pujan. Use for 'bhoomi pujan date', 'foundation stone muhurta', 'which corner to start digging'.",
		Title:       "Bhoomi Pujan Muhurta",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MuhurtaDateInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/muhurta/bhoomi-pujan", in.toQuery())
	})
}

func registerMuhurtaUpanayan(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "muhurta_upanayan",
		Description: "Score how auspicious a specific date (and optional time) is for the upanayan (sacred thread, janeu) ceremony at a location — returns a 0-100 score, a grade, and the factor breakdown (tithi, vara, nakshatra, Uttarayana or Dakshinayana, part of the day, Guru/Shukra asta, Rahu Kaal, Yamaganda, Durmuhurta, Bhadra, adhika masa)." + prohibitedNote("Guru or Shukra asta, adhika masa, Dakshinayana, Bhadra") + sanskarNotesNote + " Evaluates ONE date per call. Use for 'is this date good for janeu', 'upanayan muhurta check', 'sacred thread ceremony date'.",
		Title:       "Upanayan Muhurta",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MuhurtaDateInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/muhurta/upanayan", in.toQuery())
	})
}

func registerMuhurtaAnnaprashan(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "muhurta_annaprashan",
		Description: "Score how auspicious a specific date (and optional time) is for a baby's annaprashan (first solid food) at a location — returns a 0-100 score, a grade, and the factor breakdown (tithi, vara, nakshatra, paksha, Rahu Kaal, Yamaganda, Durmuhurta, Bhadra)." + prohibitedNote("Bhadra") + sanskarNotesNote + " With the child's birth_date and sex a note gives the running month and the months tradition names for a boy or a girl. Evaluates ONE date per call. Use for 'annaprashan date for my son', 'first rice feeding muhurta', 'is this day good for annaprashan'.",
		Title:       "Annaprashan Muhurta",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MuhurtaAnnaprashanInput) (*mcp.CallToolResult, any, error) {
		q := in.MuhurtaChildInput.toQuery()
		setIf(q, "sex", lowerTrim(in.Sex))
		return callPassthrough(ctx, c, "/v1/muhurta/annaprashan", q)
	})
}

func registerMuhurtaVidyarambh(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "muhurta_vidyarambh",
		Description: "Score how auspicious a specific date (and optional time) is for vidyarambh (a child's first letters, aksharabhyasam, start of learning) at a location — returns a 0-100 score, a grade, and the factor breakdown (tithi, vara, nakshatra, the Sun's course, Rahu Kaal, Yamaganda, Durmuhurta, Bhadra, adhika masa)." + prohibitedNote("adhika masa, Bhadra") + sanskarNotesNote + " With the child's birth_date a note gives the running year. Evaluates ONE date per call. Use for 'vidyarambh muhurta', 'aksharabhyasam date', 'good day to start my child's education'.",
		Title:       "Vidyarambh Muhurta",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MuhurtaChildInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/muhurta/vidyarambh", in.toQuery())
	})
}

func registerMuhurtaKarnavedha(s *mcp.Server, c *apiClient) {
	addLocalizedTool(s, &mcp.Tool{
		Name:        "muhurta_karnavedha",
		Description: "Score how auspicious a specific date (and optional time) is for karnavedha (the ear-piercing ceremony) at a location — returns a 0-100 score, a grade, and the factor breakdown (tithi, vara, nakshatra, solar month, Hari Shayana, Guru/Shukra asta, Rahu Kaal, Yamaganda, Durmuhurta, Bhadra, adhika masa)." + prohibitedNote("Guru or Shukra asta, adhika masa, Hari Shayana, Bhadra") + sanskarNotesNote + " With the child's birth_date a note gives the running day, month and year. Evaluates ONE date per call. Use for 'ear piercing muhurta', 'karnavedha date check', 'is this day good for ear piercing'.",
		Title:       "Karnavedha Muhurta",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in MuhurtaChildInput) (*mcp.CallToolResult, any, error) {
		return callPassthrough(ctx, c, "/v1/muhurta/karnavedha", in.toQuery())
	})
}
