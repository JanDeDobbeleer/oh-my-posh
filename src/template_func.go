package main

import (
	"errors"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var funcMapCache map[string]interface{}

func funcMap() template.FuncMap {
	if funcMapCache != nil {
		return template.FuncMap(funcMapCache)
	}
	funcMapCache = map[string]interface{}{
		"secondsRound": secondsRound,
	}
	for key, fun := range sprig.TxtFuncMap() {
		if _, ok := funcMapCache[key]; !ok {
			funcMapCache[key] = fun
		}
	}
	return template.FuncMap(funcMapCache)
}

func parseSeconds(seconds interface{}) (int, error) {
	switch seconds := seconds.(type) {
	default:
		return 0, errors.New("invalid seconds type")
	case string:
		return strconv.Atoi(seconds)
	case int:
		return seconds, nil
	case int64:
		return int(seconds), nil
	case float64:
		return int(seconds), nil
	}
}

func secondsRound(seconds interface{}) string {
	s, err := parseSeconds(seconds)
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
		hour   = minute * 60
		day    = hour * 24
		month  = day * 30
		year   = day * 365
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
