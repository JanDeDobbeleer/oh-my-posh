package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// Not logged as an error — hide_outside_ramadan being true and it not being Ramadan is expected, not a failure.
var errNotRamadan = errors.New("not in Ramadan")

// Sehar is Fajr, Iftar is Maghrib.
type Ramadan struct {
	Base
	Fajr          string
	Iftar         string
	Imsak         string
	NextEvent     string
	TimeRemaining string
	RozaNumber    int
	Fasting       bool
}

const (
	RamadanLatitude  options.Option = "latitude"
	RamadanLongitude options.Option = "longitude"
	RamadanCity      options.Option = "city"
	RamadanCountry   options.Option = "country"
	// 0-23, default 3 = Muslim World League.
	RamadanMethod options.Option = "method"
	// 0=Shafi, 1=Hanafi.
	RamadanSchool      options.Option = "school"
	RamadanHideOutside options.Option = "hide_outside_ramadan"
	// Overrides the first day of Ramadan for local moon sighting differences.
	RamadanFirstRozaDate options.Option = "first_roza_date"
)

type ramadanTimings struct {
	Fajr    string `json:"Fajr"`
	Imsak   string `json:"Imsak"`
	Maghrib string `json:"Maghrib"`
}

type ramadanHijriMonth struct {
	Number int `json:"number"`
}

type ramadanHijriDate struct {
	Day   string            `json:"day"`
	Month ramadanHijriMonth `json:"month"`
}

type ramadanDate struct {
	Hijri ramadanHijriDate `json:"hijri"`
}

type ramadanData struct {
	Timings ramadanTimings `json:"timings"`
	Date    ramadanDate    `json:"date"`
}

type ramadanResponse struct {
	Data ramadanData `json:"data"`
}

func (r *Ramadan) Template() string {
	return " \U0001F319 Roza {{.RozaNumber}} \u00b7 {{.NextEvent}} in {{.TimeRemaining}} "
}

func (r *Ramadan) Enabled() bool {
	err := r.setData()
	if err != nil {
		if !errors.Is(err, errNotRamadan) {
			log.Error(err)
		}
		return false
	}

	return true
}

func (r *Ramadan) setData() error {
	now := time.Now()
	date := now.Format("02-01-2006")

	apiURL, err := r.buildURL(date)
	if err != nil {
		return err
	}

	httpTimeout := r.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	body, err := r.env.HTTPRequest(apiURL, nil, httpTimeout)
	if err != nil {
		return err
	}

	var response ramadanResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return err
	}

	data := response.Data

	// Determine if we are currently in Ramadan and compute the roza number.
	firstRozaStr := r.options.String(RamadanFirstRozaDate, "")
	hideOutside := r.options.Bool(RamadanHideOutside, true)

	inRamadan, rozaNumber := r.resolveRamadanDay(now, data, firstRozaStr)

	if !inRamadan && hideOutside {
		return errNotRamadan
	}

	if inRamadan {
		r.RozaNumber = rozaNumber
	}

	fajrTime, err := parseEventTime(now, data.Timings.Fajr)
	if err != nil {
		return fmt.Errorf("failed to parse Fajr time: %w", err)
	}

	iftarTime, err := parseEventTime(now, data.Timings.Maghrib)
	if err != nil {
		return fmt.Errorf("failed to parse Iftar time: %w", err)
	}

	imsakTime, err := parseEventTime(now, data.Timings.Imsak)
	if err != nil {
		return fmt.Errorf("failed to parse Imsak time: %w", err)
	}

	r.Fajr = fajrTime.Format("15:04")
	r.Iftar = iftarTime.Format("15:04")
	r.Imsak = imsakTime.Format("15:04")

	// When past Iftar, fetch tomorrow's Fajr from the API for a DST-accurate countdown.
	// Falls back to the same wall-clock time on the next calendar day if the fetch fails.
	var tomorrowFajr time.Time
	if !now.Before(iftarTime) {
		tomorrow := now.AddDate(0, 0, 1)
		var fetchErr error
		tomorrowFajr, fetchErr = r.fetchFajrTime(tomorrow)
		if fetchErr != nil {
			tomorrowFajr = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
				fajrTime.Hour(), fajrTime.Minute(), 0, 0, fajrTime.Location())
		}
	}

	r.computeNextEvent(now, fajrTime, iftarTime, tomorrowFajr)

	return nil
}

// tomorrowFajrTime must be populated by the caller when now is past Iftar; it is ignored otherwise.
func (r *Ramadan) computeNextEvent(now, fajrTime, iftarTime, tomorrowFajrTime time.Time) {
	r.Fasting = !now.Before(fajrTime) && now.Before(iftarTime)

	if now.Before(fajrTime) {
		r.NextEvent = "Sehar"
		r.TimeRemaining = formatDuration(fajrTime.Sub(now))
		return
	}

	if now.Before(iftarTime) {
		r.NextEvent = "Iftar"
		r.TimeRemaining = formatDuration(iftarTime.Sub(now))
		return
	}

	// After Iftar — use tomorrow's Fajr time fetched from the API (or an AddDate fallback).
	r.NextEvent = "Sehar"
	r.TimeRemaining = formatDuration(tomorrowFajrTime.Sub(now))
}

func (r *Ramadan) fetchFajrTime(date time.Time) (time.Time, error) {
	dateStr := date.Format("02-01-2006")

	apiURL, err := r.buildURL(dateStr)
	if err != nil {
		return time.Time{}, err
	}

	httpTimeout := r.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	body, err := r.env.HTTPRequest(apiURL, nil, httpTimeout)
	if err != nil {
		return time.Time{}, err
	}

	var response ramadanResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return time.Time{}, err
	}

	return parseEventTime(date, response.Data.Timings.Fajr)
}

// City+country takes precedence over lat/lng when both are provided.
func (r *Ramadan) buildURL(date string) (string, error) {
	method := r.options.Int(RamadanMethod, 3)
	school := r.options.Int(RamadanSchool, 0)

	city := r.options.String(RamadanCity, "")
	country := r.options.String(RamadanCountry, "")

	if city != "" && country != "" {
		return fmt.Sprintf(
			"https://api.aladhan.com/v1/timingsByCity/%s?city=%s&country=%s&method=%d&school=%d",
			date,
			url.QueryEscape(city),
			url.QueryEscape(country),
			method,
			school,
		), nil
	}

	if r.options.Any(RamadanLatitude, nil) == nil || r.options.Any(RamadanLongitude, nil) == nil {
		return "", errors.New("no location configured: set city+country or latitude+longitude")
	}

	lat := r.options.Float64(RamadanLatitude, 0)
	lng := r.options.Float64(RamadanLongitude, 0)

	return fmt.Sprintf(
		"https://api.aladhan.com/v1/timings/%s?latitude=%g&longitude=%g&method=%d&school=%d",
		date, lat, lng, method, school,
	), nil
}

// Returns whether today is in Ramadan and the roza (day) number. When first_roza_date is
// set it overrides the API's Hijri month detection.
func (r *Ramadan) resolveRamadanDay(now time.Time, data ramadanData, firstRozaStr string) (bool, int) {
	if firstRozaStr != "" {
		firstRoza, err := time.ParseInLocation("2006-01-02", firstRozaStr, now.Location())
		if err == nil {
			// Use UTC noon arithmetic to avoid DST off-by-one errors.
			ny, nm, nd := now.Date()
			fy, fm, fd := firstRoza.Date()
			nowDay := time.Date(ny, nm, nd, 12, 0, 0, 0, time.UTC)
			firstDay := time.Date(fy, fm, fd, 12, 0, 0, 0, time.UTC)

			daysSince := int(nowDay.Sub(firstDay).Hours() / 24)
			if daysSince >= 0 && daysSince < 30 {
				return true, daysSince + 1
			}

			return false, 0
		}
		// Parse error: fall through to API-based Hijri month detection.
	}

	if data.Date.Hijri.Month.Number != 9 {
		return false, 0
	}

	rozaNumber := 0
	if _, err := fmt.Sscanf(data.Date.Hijri.Day, "%d", &rozaNumber); err != nil {
		return false, 0
	}

	return true, rozaNumber
}

func parseEventTime(now time.Time, hhmm string) (time.Time, error) {
	// The API may return timezone-suffixed values like "05:23 (PKT)"; strip any suffix.
	timeStr := hhmm
	if len(timeStr) > 5 {
		timeStr = timeStr[:5]
	}

	parsed, err := time.ParseInLocation("15:04", timeStr, now.Location())
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location()), nil
}

// Formats as "Xh Ym", or just "Ym" when under an hour.
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	totalMinutes := int(d.Minutes())
	h := totalMinutes / 60
	m := totalMinutes % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}

	return fmt.Sprintf("%dm", m)
}
