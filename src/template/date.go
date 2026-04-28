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
		if epoch, err := strconv.ParseInt(v, 10, 64); err == nil {
			t = time.Unix(epoch, 0)
		} else {
			t = time.Now()
		}
	default:
		t = time.Now()
	}

	loc, err := time.LoadLocation(zone)
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}

	return t.In(loc).Format(fmt)
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
