package template

import (
	"strings"
	"sync"
)

const (
	defaultDateLayout = "2006-01-02"
	defaultTimeLayout = "15:04"
)

var (
	localeOnce       sync.Once
	localeDateLayout string
	localeTimeLayout string

	// localeLayoutsResolver is set by init() in the platform-specific files
	// (locale_windows.go / locale_unix.go). It is never called until the first
	// time a template actually uses localeShortDate or localeShortTime, so the
	// OS is not queried on every prompt render — only when these functions are
	// needed, and then the results are cached for the lifetime of the process.
	localeLayoutsResolver func() (dateLayout, timeLayout string)
)

// getLocaleLayouts returns the OS short-date and short-time layouts as Go
// time format strings. The OS is queried at most once per process.
func getLocaleLayouts() (string, string) {
	localeOnce.Do(func() {
		if localeLayoutsResolver != nil {
			localeDateLayout, localeTimeLayout = localeLayoutsResolver()
		}
	})

	dateLayout := localeDateLayout
	if dateLayout == "" {
		dateLayout = defaultDateLayout
	}

	timeLayout := localeTimeLayout
	if timeLayout == "" {
		timeLayout = defaultTimeLayout
	}

	return dateLayout, timeLayout
}

// localeShortDate formats date using the OS short-date regional setting.
//
// Example:
//
//	{{ localeShortDate .SomeTime }}
func localeShortDate(date any) string {
	layout, _ := getLocaleLayouts()
	return dateInZone(layout, date, "Local")
}

// localeShortTime formats time using the OS short-time regional setting.
//
// Example:
//
//	{{ localeShortTime .SomeTime }}
func localeShortTime(date any) string {
	_, layout := getLocaleLayouts()
	return dateInZone(layout, date, "Local")
}

// windowsPatternToGoLayout converts a Windows date/time format string
// (as stored in HKCU\Control Panel\International) to a Go time layout string.
//
// Tokens are replaced longest-match-first so that e.g. "yyyy" is not
// partially consumed as "yy".
func windowsPatternToGoLayout(pattern string) string {
	// Order matters: longer tokens must be replaced before shorter ones.
	replacements := []struct{ from, to string }{
		{"dddd", "Monday"},
		{"ddd", "Mon"},
		{"dd", "02"},
		{"d", "2"},
		{"MMMM", "January"},
		{"MMM", "Jan"},
		{"MM", "01"},
		{"M", "1"},
		{"yyyy", "2006"},
		{"yy", "06"},
		// 24-hour (H/HH) — Go has no unpadded 24-hour; both map to "15".
		{"HH", "15"},
		{"H", "15"},
		// 12-hour
		{"hh", "03"},
		{"h", "3"},
		// Minutes
		{"mm", "04"},
		{"m", "4"},
		// Seconds
		{"ss", "05"},
		{"s", "5"},
		// AM/PM — Go has no single-character AM/PM token; both "t" and "tt" map to "PM".
		// "t" will always render as two characters ("AM"/"PM"), matching "tt" behaviour.
		{"tt", "PM"},
		{"t", "PM"},
	}

	// Walk through the pattern rune-by-rune, matching the longest token at
	// each position so that literal characters (separators like "/" and ":")
	// are passed through unchanged.
	var out strings.Builder
	i := 0
	for i < len(pattern) {
		matched := false
		for _, r := range replacements {
			if strings.HasPrefix(pattern[i:], r.from) {
				out.WriteString(r.to)
				i += len(r.from)
				matched = true
				break
			}
		}
		if !matched {
			out.WriteByte(pattern[i])
			i++
		}
	}

	return out.String()
}

// posixPatternToGoLayout converts a POSIX strftime format string (as returned
// by `locale -k LC_TIME` d_fmt / t_fmt) to a Go time layout string.
func posixPatternToGoLayout(pattern string) string {
	replacements := map[string]string{
		"%Y": "2006",
		"%y": "06",
		"%m": "01",
		"%d": "02",
		"%e": "2", // space-padded day; Go has no space padding — use unpadded
		"%H": "15",
		"%I": "03",
		"%M": "04",
		"%S": "05",
		"%p": "PM",
		"%P": "pm",
		"%A": "Monday",
		"%a": "Mon",
		"%B": "January",
		"%b": "Jan",
		"%Z": "MST",
		"%z": "-0700",
		"%%": "%",
	}

	var out strings.Builder
	i := 0
	for i < len(pattern) {
		if i+1 < len(pattern) && pattern[i] == '%' {
			token := pattern[i : i+2]
			if rep, ok := replacements[token]; ok {
				out.WriteString(rep)
				i += 2
				continue
			}
		}
		out.WriteByte(pattern[i])
		i++
	}

	return out.String()
}
