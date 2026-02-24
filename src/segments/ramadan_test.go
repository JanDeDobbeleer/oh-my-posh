package segments

import (
	"errors"
	"testing"
	libtime "time"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

const ramadanTestResponse = `{
  "code": 200,
  "status": "OK",
  "data": {
    "timings": {
      "Fajr": "05:15",
      "Imsak": "05:05",
      "Maghrib": "18:30"
    },
    "date": {
      "hijri": {
        "day": "5",
        "month": { "number": 9 }
      }
    }
  }
}`

const ramadanNonRamadanResponse = `{
  "code": 200,
  "status": "OK",
  "data": {
    "timings": {
      "Fajr": "05:15",
      "Imsak": "05:05",
      "Maghrib": "18:30"
    },
    "date": {
      "hijri": {
        "day": "10",
        "month": { "number": 7 }
      }
    }
  }
}`

func TestRamadanSegment(t *testing.T) {
	today := libtime.Now()
	date := today.Format("02-01-2006")
	// firstRoza 5 days before today so roza number is always 6 regardless of when tests run
	firstRoza := today.AddDate(0, 0, -5).Format("2006-01-02")

	cases := []struct {
		APIError        error
		Props           options.Map
		Case            string
		APIResponse     string
		ExpectedRoza    int
		ExpectedEnabled bool
	}{
		{
			Case:        "in Ramadan via API hijri month",
			APIResponse: ramadanTestResponse,
			Props: options.Map{
				RamadanLatitude:  51.5,
				RamadanLongitude: -0.1,
			},
			ExpectedEnabled: true,
			ExpectedRoza:    5,
		},
		{
			Case:        "in Ramadan via first_roza_date override",
			APIResponse: ramadanNonRamadanResponse,
			Props: options.Map{
				RamadanCity:          "Lahore",
				RamadanCountry:       "Pakistan",
				RamadanFirstRozaDate: firstRoza,
			},
			ExpectedEnabled: true,
			ExpectedRoza:    6,
		},
		{
			Case:        "not in Ramadan, hide=true (default)",
			APIResponse: ramadanNonRamadanResponse,
			Props: options.Map{
				RamadanLatitude:  51.5,
				RamadanLongitude: -0.1,
			},
			ExpectedEnabled: false,
		},
		{
			Case:        "not in Ramadan, hide=false shows segment",
			APIResponse: ramadanNonRamadanResponse,
			Props: options.Map{
				RamadanLatitude:    51.5,
				RamadanLongitude:   -0.1,
				RamadanHideOutside: false,
			},
			ExpectedEnabled: true,
			ExpectedRoza:    0,
		},
		{
			Case:        "API error returns false",
			APIResponse: "",
			APIError:    errors.New("network error"),
			Props: options.Map{
				RamadanLatitude:  51.5,
				RamadanLongitude: -0.1,
			},
			ExpectedEnabled: false,
		},
		{
			Case:            "no location configured returns false",
			APIResponse:     ramadanTestResponse,
			Props:           options.Map{},
			ExpectedEnabled: false,
		},
	}

	tomorrow := today.AddDate(0, 0, 1)
	tomorrowDate := tomorrow.Format("02-01-2006")

	for _, tc := range cases {
		env := &mock.Environment{}

		// Build the expected URL based on props to pass to the mock
		city, hasCity := tc.Props[RamadanCity]
		country, hasCountry := tc.Props[RamadanCountry]
		_, hasLat := tc.Props[RamadanLatitude]
		_, hasLng := tc.Props[RamadanLongitude]

		var apiURL string
		var tomorrowAPIURL string

		switch {
		case hasCity && hasCountry:
			apiURL = "https://api.aladhan.com/v1/timingsByCity/" + date +
				"?city=" + city.(string) + "&country=" + country.(string) + "&method=3&school=0"
			tomorrowAPIURL = "https://api.aladhan.com/v1/timingsByCity/" + tomorrowDate +
				"?city=" + city.(string) + "&country=" + country.(string) + "&method=3&school=0"
		case hasLat && hasLng:
			apiURL = "https://api.aladhan.com/v1/timings/" + date +
				"?latitude=51.5&longitude=-0.1&method=3&school=0"
			tomorrowAPIURL = "https://api.aladhan.com/v1/timings/" + tomorrowDate +
				"?latitude=51.5&longitude=-0.1&method=3&school=0"
		}

		if apiURL != "" {
			env.On("HTTPRequest", apiURL).Return([]byte(tc.APIResponse), tc.APIError)
		}

		// Also mock tomorrow's URL (called after Iftar for the next Sehar countdown).
		// Marked as Maybe() so the test passes regardless of what time of day it runs.
		if tomorrowAPIURL != "" {
			env.On("HTTPRequest", tomorrowAPIURL).Maybe().Return([]byte(tc.APIResponse), nil)
		}

		r := &Ramadan{}
		r.Init(tc.Props, env)

		enabled := r.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedRoza, r.RozaNumber, tc.Case)
	}
}

func TestComputeNextEvent(t *testing.T) {
	base := libtime.Date(2026, 3, 1, 0, 0, 0, 0, libtime.UTC)
	fajr := base.Add(5*libtime.Hour + 15*libtime.Minute)
	iftar := base.Add(18*libtime.Hour + 30*libtime.Minute)
	nextDay := fajr.AddDate(0, 0, 1)
	tomorrowFajr := libtime.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), fajr.Hour(), fajr.Minute(), 0, 0, fajr.Location())

	cases := []struct {
		Now               libtime.Time
		Case              string
		ExpectedNextEvent string
		ExpectedFasting   bool
	}{
		{Case: "before fajr", Now: fajr.Add(-10 * libtime.Minute), ExpectedFasting: false, ExpectedNextEvent: "Sehar"},
		{Case: "exactly at fajr (boundary)", Now: fajr, ExpectedFasting: true, ExpectedNextEvent: "Iftar"},
		{Case: "after fajr, before iftar", Now: fajr.Add(1 * libtime.Hour), ExpectedFasting: true, ExpectedNextEvent: "Iftar"},
		{Case: "exactly at iftar", Now: iftar, ExpectedFasting: false, ExpectedNextEvent: "Sehar"},
		{Case: "after iftar", Now: iftar.Add(30 * libtime.Minute), ExpectedFasting: false, ExpectedNextEvent: "Sehar"},
	}

	for _, tc := range cases {
		r := &Ramadan{}
		r.computeNextEvent(tc.Now, fajr, iftar, tomorrowFajr)
		assert.Equal(t, tc.ExpectedFasting, r.Fasting, tc.Case)
		assert.Equal(t, tc.ExpectedNextEvent, r.NextEvent, tc.Case)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Input    libtime.Duration
	}{
		{"hours and minutes", "3h 42m", 3*libtime.Hour + 42*libtime.Minute},
		{"minutes only", "25m", 25 * libtime.Minute},
		{"zero", "0m", 0},
		{"negative clamped to zero", "0m", -5 * libtime.Minute},
		{"exact hour", "1h 0m", libtime.Hour},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.Expected, formatDuration(tc.Input), tc.Case)
	}
}

func TestParseEventTime(t *testing.T) {
	now := libtime.Date(2026, 2, 24, 12, 0, 0, 0, libtime.UTC)

	cases := []struct {
		Case           string
		Input          string
		ExpectedHour   int
		ExpectedMinute int
	}{
		{"plain HH:MM", "05:23", 5, 23},
		{"with timezone suffix", "05:23 (PKT)", 5, 23},
		{"midnight", "00:04", 0, 4},
	}

	for _, tc := range cases {
		result, err := parseEventTime(now, tc.Input)
		assert.NoError(t, err, tc.Case)
		assert.Equal(t, tc.ExpectedHour, result.Hour(), tc.Case)
		assert.Equal(t, tc.ExpectedMinute, result.Minute(), tc.Case)
		assert.Equal(t, 2026, result.Year(), tc.Case)
	}
}

func TestResolveRamadanDay(t *testing.T) {
	now := libtime.Date(2026, 2, 24, 12, 0, 0, 0, libtime.UTC)
	r := &Ramadan{}

	ramadanData := ramadanData{
		Date: ramadanDate{
			Hijri: ramadanHijriDate{
				Day:   "5",
				Month: ramadanHijriMonth{Number: 9},
			},
		},
	}
	nonRamadanData := ramadanData
	nonRamadanData.Date.Hijri.Month.Number = 7

	// via hijri month
	inRamadan, roza := r.resolveRamadanDay(now, ramadanData, "")
	assert.True(t, inRamadan)
	assert.Equal(t, 5, roza)

	// not Ramadan month
	inRamadan, roza = r.resolveRamadanDay(now, nonRamadanData, "")
	assert.False(t, inRamadan)
	assert.Equal(t, 0, roza)

	// first_roza_date override — today is day 6 (2026-02-24, first roza 2026-02-19)
	inRamadan, roza = r.resolveRamadanDay(now, nonRamadanData, "2026-02-19")
	assert.True(t, inRamadan)
	assert.Equal(t, 6, roza)

	// first_roza_date override — date before start
	beforeStart := libtime.Date(2026, 2, 18, 12, 0, 0, 0, libtime.UTC)
	inRamadan, _ = r.resolveRamadanDay(beforeStart, nonRamadanData, "2026-02-19")
	assert.False(t, inRamadan)

	// first_roza_date override — past 30 days
	afterEnd := libtime.Date(2026, 3, 22, 12, 0, 0, 0, libtime.UTC)
	inRamadan, _ = r.resolveRamadanDay(afterEnd, nonRamadanData, "2026-02-19")
	assert.False(t, inRamadan)
}

func TestRamadanComputeNextEvent(t *testing.T) {
	fajr := libtime.Date(2026, 3, 10, 5, 15, 0, 0, libtime.UTC)
	iftar := libtime.Date(2026, 3, 10, 18, 30, 0, 0, libtime.UTC)

	cases := []struct {
		Case              string
		Now               libtime.Time
		ExpectedNextEvent string
		ExpectedFasting   bool
	}{
		{
			Case:              "before Fajr",
			Now:               libtime.Date(2026, 3, 10, 4, 0, 0, 0, libtime.UTC),
			ExpectedNextEvent: "Sehar",
			ExpectedFasting:   false,
		},
		{
			Case:              "during fasting window",
			Now:               libtime.Date(2026, 3, 10, 12, 0, 0, 0, libtime.UTC),
			ExpectedNextEvent: "Iftar",
			ExpectedFasting:   true,
		},
		{
			Case:              "after Iftar",
			Now:               libtime.Date(2026, 3, 10, 20, 0, 0, 0, libtime.UTC),
			ExpectedNextEvent: "Sehar",
			ExpectedFasting:   false,
		},
	}

	nextDay := fajr.AddDate(0, 0, 1)
	tomorrowFajr := libtime.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), fajr.Hour(), fajr.Minute(), 0, 0, fajr.Location())

	for _, tc := range cases {
		r := &Ramadan{}
		r.computeNextEvent(tc.Now, fajr, iftar, tomorrowFajr)
		assert.Equal(t, tc.ExpectedNextEvent, r.NextEvent, tc.Case)
		assert.Equal(t, tc.ExpectedFasting, r.Fasting, tc.Case)
		assert.NotEmpty(t, r.TimeRemaining, tc.Case)
	}
}
