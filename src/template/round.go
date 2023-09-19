package template

import (
	"strconv"
	"strings"
)

func secondsRound(seconds any) string {
	s, err := toInt(seconds)
	if err != nil {
		return err.Error()
	}
	if s == 0 {
		return "0s"
	}
	neg := s < 0
	if neg {
		s = -s
	}

	var (
		second = 1
		minute = 60
		hour   = 3600
		day    = 86400
		month  = 2629800
		year   = 31560000
	)
	var builder strings.Builder
	writePart := func(unit int, name string) {
		if s >= unit {
			builder.WriteString(" ")
			builder.WriteString(strconv.Itoa(s / unit))
			builder.WriteString(name)
			s %= unit
		}
	}
	writePart(year, "y")
	writePart(month, "mo")
	writePart(day, "d")
	writePart(hour, "h")
	writePart(minute, "m")
	writePart(second, "s")
	return strings.Trim(builder.String(), " ")
}
