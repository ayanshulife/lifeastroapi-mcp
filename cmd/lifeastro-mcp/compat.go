package main

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// This file holds the helpers that keep tool inputs aligned with what the
// API actually accepts. They exist because of the 2026-10 live sweep
// (live_test.go): the API requires an explicit date on every "what about
// today" endpoint and reads natal data under a "birth." prefix on every
// endpoint that also takes a query/transit moment. Tools let the assistant
// omit "today", so the MCP layer fills it in.

// zoneOr loads an IANA zone, falling back to UTC for an empty or unknown name.
func zoneOr(tz string) *time.Location {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	return time.UTC
}

// dateOrToday returns d, or today's date (YYYY-MM-DD) in tz when d is empty.
func dateOrToday(d, tz string) string {
	if d = strings.TrimSpace(d); d != "" {
		return d
	}
	return time.Now().In(zoneOr(tz)).Format("2006-01-02")
}

// timeOrNow returns t, or the current wall-clock time (HH:MM) in tz when t is empty.
func timeOrNow(t, tz string) string {
	if t = strings.TrimSpace(t); t != "" {
		return t
	}
	return time.Now().In(zoneOr(tz)).Format("15:04")
}

// tzOrUTC returns tz, or "UTC" when it is empty. The API has no default zone.
func tzOrUTC(tz string) string {
	if tz = strings.TrimSpace(tz); tz != "" {
		return tz
	}
	return "UTC"
}

// dualQuery builds the query for endpoints that need a natal chart AND a
// query moment (transits, sade sati, tara bala, dual-moment narratives):
// natal data under "birth.*", the moment as unprefixed date/tz.
//
// The birth place is also sent unprefixed as lat/lon: several handlers
// (narrative horoscopes, transit ashtakavarga, varshaphal themes) read the
// natal location that way and reject the request without it.
//
// Sending natal data unprefixed instead is not a harmless variation — the
// API then treats the birth date as the transit date and answers 200 with
// numbers for the wrong day.
func dualQuery(b BirthInput, onDate, onTz string) url.Values {
	q := url.Values{}
	lat := strconv.FormatFloat(b.Lat, 'f', -1, 64)
	lon := strconv.FormatFloat(b.Lon, 'f', -1, 64)
	q.Set("birth.lat", lat)
	q.Set("birth.lon", lon)
	q.Set("birth.date", b.Date)
	q.Set("birth.time", b.Time)
	q.Set("birth.tz", b.Tz)
	q.Set("lat", lat)
	q.Set("lon", lon)
	tz := strings.TrimSpace(onTz)
	if tz == "" {
		tz = b.Tz
	}
	q.Set("tz", tzOrUTC(tz))
	q.Set("date", dateOrToday(onDate, tz))
	return q
}

// vimshottariLords maps the spellings an assistant is likely to send for a
// Vimshottari dasha lord to the name the API's path segments accept.
var vimshottariLords = map[string]string{
	"sun": "Sun", "su": "Sun", "surya": "Sun",
	"moon": "Moon", "mo": "Moon", "chandra": "Moon",
	"mars": "Mars", "ma": "Mars", "mangal": "Mars",
	"mercury": "Mercury", "mer": "Mercury", "me": "Mercury", "budh": "Mercury", "budha": "Mercury",
	"jupiter": "Jupiter", "jup": "Jupiter", "ju": "Jupiter", "guru": "Jupiter",
	"venus": "Venus", "ven": "Venus", "ve": "Venus", "shukra": "Venus",
	"saturn": "Saturn", "sat": "Saturn", "sa": "Saturn", "shani": "Saturn",
	"rahu": "Rahu", "rah": "Rahu", "ra": "Rahu",
	"ketu": "Ketu", "ket": "Ketu", "ke": "Ketu",
}

// dashaLord normalises a dasha-lord name for use as a URL path segment.
// Unknown input is passed through (escaped) so the API's own error, which
// lists the valid names, reaches the assistant.
func dashaLord(s string) string {
	if v, ok := vimshottariLords[strings.ToLower(strings.TrimSpace(s))]; ok {
		return v
	}
	return url.PathEscape(strings.TrimSpace(s))
}

// yearOrCurrent returns y, or the current calendar year when y is zero.
func yearOrCurrent(y int) int {
	if y != 0 {
		return y
	}
	return time.Now().Year()
}

// yearMonth formats a year and month as the "YYYY-MM" the API expects.
func yearMonth(y, m int) string {
	return strconv.Itoa(y) + "-" + twoDigits(m)
}

func twoDigits(n int) string {
	if n >= 0 && n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// addYears returns date (YYYY-MM-DD) shifted by n years; an unparseable
// date is returned unchanged so the API reports the format problem.
func addYears(date string, n int) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.AddDate(n, 0, 0).Format("2006-01-02")
}

// westernBirthPrefixQuery emits a Western natal chart under the "birth."
// prefix, for Western endpoints that also take a transit moment or window.
func westernBirthPrefixQuery(b WesternBirthInput) url.Values {
	q := url.Values{}
	q.Set("birth.lat", strconv.FormatFloat(b.Lat, 'f', -1, 64))
	q.Set("birth.lon", strconv.FormatFloat(b.Lon, 'f', -1, 64))
	q.Set("birth.date", b.Date)
	q.Set("birth.time", b.Time)
	q.Set("birth.tz", b.Tz)
	if b.HouseSystem != "" {
		q.Set("house_system", b.HouseSystem)
	}
	return q
}
