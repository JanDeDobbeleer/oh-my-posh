package template

import (
	"sync"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

// resetLocaleCache resets the locale sync.Once so individual test cases can
// inject their own layouts without cross-contamination.
func resetLocaleCache(dateLayout, timeLayout string) {
	localeOnce = sync.Once{}
	localeDateLayout = dateLayout
	localeTimeLayout = timeLayout
}

// TestWindowsPatternToGoLayout verifies the Windows date/time format → Go layout converter.
//
//nolint:dupl
func TestWindowsPatternToGoLayout(t *testing.T) {
	cases := []struct {
		Case     string
		Input    string
		Expected string
	}{
		{Case: "ISO-style date", Input: "yyyy-MM-dd", Expected: defaultDateLayout},
		{Case: "US short date", Input: "M/d/yyyy", Expected: "1/2/2006"},
		{Case: "EU date with dots", Input: "dd.MM.yyyy", Expected: "02.01.2006"},
		{Case: "24-hour time", Input: "HH:mm", Expected: defaultTimeLayout},
		{Case: "12-hour time with AM/PM", Input: "h:mm tt", Expected: "3:04 PM"},
		{Case: "zero-padded 12-hour", Input: "hh:mm tt", Expected: "03:04 PM"},
		{Case: "abbreviated day name", Input: "ddd, MMM d yyyy", Expected: "Mon, Jan 2 2006"},
		{Case: "full day and month name", Input: "dddd, MMMM d, yyyy", Expected: "Monday, January 2, 2006"},
		{Case: "2-digit year", Input: "dd/MM/yy", Expected: "02/01/06"},
		{Case: "with seconds", Input: "HH:mm:ss", Expected: "15:04:05"},
	}

	for _, tc := range cases {
		got := windowsPatternToGoLayout(tc.Input)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

// TestPosixPatternToGoLayout verifies the POSIX strftime → Go layout converter.
//
//nolint:dupl
func TestPosixPatternToGoLayout(t *testing.T) {
	cases := []struct {
		Case     string
		Input    string
		Expected string
	}{
		{Case: "ISO date", Input: "%Y-%m-%d", Expected: defaultDateLayout},
		{Case: "US date", Input: "%m/%d/%Y", Expected: "01/02/2006"},
		{Case: "EU date with dots", Input: "%d.%m.%Y", Expected: "02.01.2006"},
		{Case: "24-hour time", Input: "%H:%M", Expected: defaultTimeLayout},
		{Case: "12-hour with AM/PM", Input: "%I:%M %p", Expected: "03:04 PM"},
		{Case: "abbreviated month", Input: "%b %e, %Y", Expected: "Jan 2, 2006"},
		{Case: "full day and month", Input: "%A, %B %e, %Y", Expected: "Monday, January 2, 2006"},
		{Case: "escaped percent", Input: "100%%", Expected: "100%"},
		{Case: "lowercase am/pm", Input: "%I:%M %P", Expected: "03:04 pm"},
		{Case: "with seconds", Input: "%H:%M:%S", Expected: "15:04:05"},
	}

	for _, tc := range cases {
		got := posixPatternToGoLayout(tc.Input)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

// TestLocaleShortDateFallback verifies that a zero-value resolver still
// produces ISO 8601 output rather than an empty string.
func TestLocaleShortDateFallback(t *testing.T) {
	mockEnv := new(mock.Environment)
	mockEnv.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(mockEnv, nil, nil)

	// Override the resolver to return empty strings (simulates resolver failure).
	localeLayoutsResolver = func() (string, string) { return "", "" }
	resetLocaleCache("", "")

	time.Local = time.UTC
	defer func() { time.Local = time.FixedZone("UTC", 0) }()

	tmpl := `{{ localeShortDate .T }}`
	ctx := struct{ T time.Time }{T: time.Unix(0, 0).UTC()}
	got, err := Render(tmpl, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "1970-01-01", got)
}

// TestLocaleShortTimeFallback verifies that a zero-value resolver still
// produces 24-hour output rather than an empty string.
func TestLocaleShortTimeFallback(t *testing.T) {
	mockEnv := new(mock.Environment)
	mockEnv.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(mockEnv, nil, nil)

	localeLayoutsResolver = func() (string, string) { return "", "" }
	resetLocaleCache("", "")

	time.Local = time.UTC
	defer func() { time.Local = time.FixedZone("UTC", 0) }()

	tmpl := `{{ localeShortTime .T }}`
	ctx := struct{ T time.Time }{T: time.Unix(0, 0).UTC()}
	got, err := Render(tmpl, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "00:00", got)
}

// TestLocaleShortDateWithISOLayout exercises the full template stack with a
// known locale layout so we can verify the output without touching the OS.
func TestLocaleShortDateWithISOLayout(t *testing.T) {
	mockEnv := new(mock.Environment)
	mockEnv.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(mockEnv, nil, nil)

	localeLayoutsResolver = func() (string, string) { return defaultDateLayout, defaultTimeLayout }
	resetLocaleCache("", "")

	time.Local = time.UTC
	defer func() { time.Local = time.FixedZone("UTC", 0) }()

	// knownEpoch = 2019-06-13 20:39:39 UTC (from date_test.go)
	tmpl := `{{ localeShortDate .T }}`
	ctx := struct{ T time.Time }{T: time.Unix(knownEpoch, 0).UTC()}
	got, err := Render(tmpl, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2019-06-13", got)
}

// TestLocaleShortTimeWith24hLayout exercises the time function with a known layout.
func TestLocaleShortTimeWith24hLayout(t *testing.T) {
	mockEnv := new(mock.Environment)
	mockEnv.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(mockEnv, nil, nil)

	localeLayoutsResolver = func() (string, string) { return defaultDateLayout, defaultTimeLayout }
	resetLocaleCache("", "")

	time.Local = time.UTC
	defer func() { time.Local = time.FixedZone("UTC", 0) }()

	tmpl := `{{ localeShortTime .T }}`
	ctx := struct{ T time.Time }{T: time.Unix(knownEpoch, 0).UTC()}
	got, err := Render(tmpl, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "20:39", got)
}

// TestLocaleShortDateWith12hUSLayout verifies a US-style locale layout.
func TestLocaleShortDateWith12hUSLayout(t *testing.T) {
	mockEnv := new(mock.Environment)
	mockEnv.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(mockEnv, nil, nil)

	localeLayoutsResolver = func() (string, string) { return "1/2/2006", "3:04 PM" }
	resetLocaleCache("", "")

	time.Local = time.UTC
	defer func() { time.Local = time.FixedZone("UTC", 0) }()

	tmpl := `{{ localeShortDate .T }} {{ localeShortTime .T }}`
	ctx := struct{ T time.Time }{T: time.Unix(knownEpoch, 0).UTC()}
	got, err := Render(tmpl, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "6/13/2019 8:39 PM", got)
}
