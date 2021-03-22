package main

import (
	"fmt"
	"strconv"

	lang "golang.org/x/text/language"
	"golang.org/x/text/message"
)

type executiontime struct {
	props  *properties
	env    environmentInfo
	output string
}

// DurationStyle how to display the time
type DurationStyle string

const (
	// ThresholdProperty represents minimum duration (milliseconds) required to enable this segment
	ThresholdProperty Property = "threshold"
	// Austin milliseconds short
	Austin DurationStyle = "austin"
	// Roundrock milliseconds long
	Roundrock DurationStyle = "roundrock"
	// Dallas milliseconds full
	Dallas DurationStyle = "dallas"
	// Galveston hour
	Galveston DurationStyle = "galveston"
	// Houston hour and milliseconds
	Houston DurationStyle = "houston"
	// Amarillo seconds
	Amarillo DurationStyle = "amarillo"
	// Round will round the output of the format
	Round DurationStyle = "round"

	second           = 1000
	minute           = 60000
	hour             = 3600000
	day              = 86400000
	secondsPerMinute = 60
	minutesPerHour   = 60
	hoursPerDay      = 24
)

func (t *executiontime) enabled() bool {
	alwaysEnabled := t.props.getBool(AlwaysEnabled, false)
	executionTimeMs := t.env.executionTime()
	thresholdMs := t.props.getFloat64(ThresholdProperty, float64(500))
	if !alwaysEnabled && executionTimeMs < thresholdMs {
		return false
	}
	style := DurationStyle(t.props.getString(Style, string(Austin)))
	t.output = t.formatDuration(int64(executionTimeMs), style)

	return t.output != ""
}

func (t *executiontime) string() string {
	return t.output
}

func (t *executiontime) init(props *properties, env environmentInfo) {
	t.props = props
	t.env = env
}

func (t *executiontime) formatDuration(ms int64, style DurationStyle) string {
	switch style {
	case Austin:
		return t.formatDurationAustin(ms)
	case Roundrock:
		return t.formatDurationRoundrock(ms)
	case Dallas:
		return t.formatDurationDallas(ms)
	case Galveston:
		return t.formatDurationGalveston(ms)
	case Houston:
		return t.formatDurationHouston(ms)
	case Amarillo:
		return t.formatDurationAmarillo(ms)
	case Round:
		return t.formatDurationRound(ms)
	default:
		return fmt.Sprintf("Style: %s is not available", style)
	}
}

func (t *executiontime) formatDurationAustin(ms int64) string {
	if ms < second {
		return fmt.Sprintf("%dms", ms%second)
	}

	seconds := float64(ms%minute) / second
	result := strconv.FormatFloat(seconds, 'f', -1, 64) + "s"

	if ms >= minute {
		result = fmt.Sprintf("%dm %s", ms/minute%secondsPerMinute, result)
	}
	if ms >= hour {
		result = fmt.Sprintf("%dh %s", ms/hour%hoursPerDay, result)
	}
	if ms >= day {
		result = fmt.Sprintf("%dd %s", ms/day, result)
	}
	return result
}

func (t *executiontime) formatDurationRoundrock(ms int64) string {
	result := fmt.Sprintf("%dms", ms%second)
	if ms >= second {
		result = fmt.Sprintf("%ds %s", ms/second%secondsPerMinute, result)
	}
	if ms >= minute {
		result = fmt.Sprintf("%dm %s", ms/minute%minutesPerHour, result)
	}
	if ms >= hour {
		result = fmt.Sprintf("%dh %s", ms/hour%hoursPerDay, result)
	}
	if ms >= day {
		result = fmt.Sprintf("%dd %s", ms/day, result)
	}
	return result
}

func (t *executiontime) formatDurationDallas(ms int64) string {
	seconds := float64(ms%minute) / second
	result := strconv.FormatFloat(seconds, 'f', -1, 64)

	if ms >= minute {
		result = fmt.Sprintf("%d:%s", ms/minute%minutesPerHour, result)
	}
	if ms >= hour {
		result = fmt.Sprintf("%d:%s", ms/hour%hoursPerDay, result)
	}
	if ms >= day {
		result = fmt.Sprintf("%d:%s", ms/day, result)
	}
	return result
}

func (t *executiontime) formatDurationGalveston(ms int64) string {
	result := fmt.Sprintf("%02d:%02d:%02d", ms/hour, ms/minute%minutesPerHour, ms%minute/second)
	return result
}

func (t *executiontime) formatDurationHouston(ms int64) string {
	milliseconds := ".0"
	if ms%second > 0 {
		// format milliseconds as a string with truncated trailing zeros
		milliseconds = strconv.FormatFloat(float64(ms%second)/second, 'f', -1, 64)
		// at this point milliseconds looks like "0.5". remove the leading "0"
		milliseconds = milliseconds[1:]
	}

	result := fmt.Sprintf("%02d:%02d:%02d%s", ms/hour, ms/minute%minutesPerHour, ms%minute/second, milliseconds)
	return result
}

func (t *executiontime) formatDurationAmarillo(ms int64) string {
	// wholeNumber represents the value to the left of the decimal point (seconds)
	wholeNumber := ms / second
	// decimalNumber represents the value to the right of the decimal point (milliseconds)
	decimalNumber := float64(ms%second) / second

	// format wholeNumber as a string with thousands separators
	printer := message.NewPrinter(lang.English)
	result := printer.Sprintf("%d", wholeNumber)

	if decimalNumber > 0 {
		// format decimalNumber as a string with truncated trailing zeros
		decimalResult := strconv.FormatFloat(decimalNumber, 'f', -1, 64)
		// at this point decimalResult looks like "0.5"
		// remove the leading "0" and append
		result += decimalResult[1:]
	}
	result += "s"

	return result
}

func (t *executiontime) formatDurationRound(ms int64) string {
	toRoundString := func(one, two int64, oneText, twoText string) string {
		if two == 0 {
			return fmt.Sprintf("%d%s", one, oneText)
		}
		return fmt.Sprintf("%d%s %d%s", one, oneText, two, twoText)
	}
	hours := ms / hour % hoursPerDay
	if ms >= day {
		return toRoundString(ms/day, hours, "d", "h")
	}
	minutes := ms / minute % secondsPerMinute
	if ms >= hour {
		return toRoundString(hours, minutes, "h", "m")
	}
	seconds := (ms % minute) / second
	if ms >= minute {
		return toRoundString(minutes, seconds, "m", "s")
	}
	if ms >= second {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dms", ms%second)
}
