package segments

import (
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type Time struct {
	Base

	CurrentDate time.Time
	Format      string
}

const (
	// TimeFormat uses the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format the current time
	TimeFormat options.Option = "time_format"
)

func (t *Time) Template() string {
	return " {{ .CurrentDate | date .Format }} "
}

func (t *Time) Enabled() bool {
	// if no date set, use now(unit testing)
	formatInput := t.options.String(TimeFormat, "15:04:05")
	t.Format = t.getTimeFormat(formatInput)
	if t.CurrentDate.IsZero() {
		t.CurrentDate = time.Now()
	}

	return true
}

var timeFormatLookup = map[string]string{
	// Maps string names to their corresponding time package constants
	"Layout":      time.Layout,      // "01/02 03:04:05PM '06 -0700"
	"ANSIC":       time.ANSIC,       // "Mon Jan _2 15:04:05 2006"
	"UnixDate":    time.UnixDate,    // "Mon Jan _2 15:04:05 MST 2006"
	"RubyDate":    time.RubyDate,    // "Mon Jan 02 15:04:05 -0700 2006"
	"RFC822":      time.RFC822,      // "02 Jan 06 15:04 MST"
	"RFC822Z":     time.RFC822Z,     // "02 Jan 06 15:04 -0700"
	"RFC850":      time.RFC850,      // "Monday, 02-Jan-06 15:04:05 MST"
	"RFC1123":     time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
	"RFC1123Z":    time.RFC1123Z,    // "Mon, 02 Jan 2006 15:04:05 -0700"
	"RFC3339":     time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
	"RFC3339Nano": time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
	"Kitchen":     time.Kitchen,     // "3:04PM"
	"Stamp":       time.Stamp,       // "Jan _2 15:04:05"
	"StampMilli":  time.StampMilli,  // "Jan _2 15:04:05.000"
	"StampMicro":  time.StampMicro,  // "Jan _2 15:04:05.000000"
	"StampNano":   time.StampNano,   // "Jan _2 15:04:05.000000000"
	"DateTime":    time.DateTime,    // "2006-01-02 15:04:05"
	"DateOnly":    time.DateOnly,    // "2006-01-02"
	"TimeOnly":    time.TimeOnly,    // "15:04:05"
}

// getTimeFormat returns the time format constant if the input matches a known format name,
// otherwise returns the input unchanged
func (t *Time) getTimeFormat(format string) string {
	if timeFormat, exists := timeFormatLookup[format]; exists {
		return timeFormat
	}
	return format
}
