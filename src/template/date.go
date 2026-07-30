package template

import (
	"strconv"
	"time"
)

// dateInZone is a replacement for sprig's dateInZone that adds support for string
// epoch values. Sprig's unixEpoch returns a string, which sprig's own date functions
// do not handle — they fall through to time.Now(). This wrapper parses numeric strings
// as Unix timestamps so that patterns like `{{ now | unixEpoch | date "..." }}` work.
func dateInZone(fmt string, date any, zone string) string {
	var t time.Time

	switch v := date.(type) {
	case time.Time:
		t = v
	case *time.Time:
		t = *v
	case int64:
		t = time.Unix(v, 0)
	case int:
		t = time.Unix(int64(v), 0)
	case int32:
		t = time.Unix(int64(v), 0)
	case string:
		t = parseDateString(v)
	default:
		t = time.Now()
	}

	loc, err := time.LoadLocation(zone)
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}

	return t.In(loc).Format(fmt)
}

// parseDateString reads the two shapes a date reaches a template as text in: a Unix epoch, which
// is what sprig's own unixEpoch hands on, and an RFC 3339 timestamp, which is how a time.Time
// marshals. The second matters for a segment rendered from recorded data rather than from its
// writer: JSON has no time type, so a recorded date arrives as the string it marshalled to, and
// falling through to time.Now() would quietly replace it with the wall clock.
func parseDateString(value string) time.Time {
	if epoch, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Unix(epoch, 0)
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed
	}

	return time.Now()
}

func ompDate(fmt string, date any) string {
	return dateInZone(fmt, date, "Local")
}

func ompDateInZone(fmt string, date any, zone string) string {
	return dateInZone(fmt, date, zone)
}

func ompHTMLDate(date any) string {
	return dateInZone("2006-01-02", date, "Local")
}

func ompHTMLDateInZone(date any, zone string) string {
	return dateInZone("2006-01-02", date, zone)
}
