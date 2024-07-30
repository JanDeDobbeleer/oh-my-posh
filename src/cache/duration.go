package cache

import (
	"strconv"

	"github.com/jandedobbeleer/oh-my-posh/src/regex"
)

type Duration string

const INFINITE = Duration("infinite")

func (d Duration) Seconds() int {
	if d == INFINITE {
		return -1
	}

	re := `(?P<AMOUNT>[0-9]*)(?P<UNIT>.*)`
	match := regex.FindNamedRegexMatch(re, string(d))
	if len(match) < 2 {
		return 0
	}

	amount := match["AMOUNT"]
	unit := match["UNIT"]

	if len(amount) == 0 {
		return 0
	}

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return 0
	}

	var multiplier int

	switch unit {
	case "second", "seconds":
		multiplier = 1
	case "minute", "minutes":
		multiplier = 60
	case "hour", "hours":
		multiplier = 3600
	case "day", "days":
		multiplier = 86400
	case "week", "weeks":
		multiplier = 604800
	case "month", "months":
		multiplier = 2592000
	}

	return amountInt * multiplier
}

func (d Duration) IsEmpty() bool {
	return len(d) == 0
}

func ToDuration(seconds int) Duration {
	if seconds == 0 {
		return ""
	}

	if seconds == -1 {
		return "infinite"
	}

	if seconds%604800 == 0 {
		return Duration(strconv.Itoa(seconds/604800) + "weeks")
	}

	if seconds%86400 == 0 {
		return Duration(strconv.Itoa(seconds/86400) + "days")
	}

	if seconds%3600 == 0 {
		return Duration(strconv.Itoa(seconds/3600) + "hours")
	}

	if seconds%60 == 0 {
		return Duration(strconv.Itoa(seconds/60) + "minutes")
	}

	return Duration(strconv.Itoa(seconds) + "seconds")
}
