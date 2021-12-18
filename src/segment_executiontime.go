package main

import (
	"fmt"
	"strconv"

	lang "golang.org/x/text/language"
	"golang.org/x/text/message"
)

type executiontime struct {
	props properties
	env   environmentInfo

	FormattedMs string
	Ms          int64
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
	t.Ms = int64(executionTimeMs)
	t.FormattedMs = t.formatDuration(style)
	return t.FormattedMs != ""
}

func (t *executiontime) string() string {
	return t.FormattedMs
}

func (t *executiontime) init(props properties, env environmentInfo) {
	t.props = props
	t.env = env
}

func (t *executiontime) formatDuration(style DurationStyle) string {
	switch style {
	case Austin:
		return t.formatDurationAustin()
	case Roundrock:
		return t.formatDurationRoundrock()
	case Dallas:
		return t.formatDurationDallas()
	case Galveston:
		return t.formatDurationGalveston()
	case Houston:
		return t.formatDurationHouston()
	case Amarillo:
		return t.formatDurationAmarillo()
	case Round:
		return t.formatDurationRound()
	default:
		return fmt.Sprintf("Style: %s is not available", style)
	}
}

func (t *executiontime) formatDurationAustin() string {
	if t.Ms < second {
		return fmt.Sprintf("%dms", t.Ms%second)
	}

	seconds := float64(t.Ms%minute) / second
	result := strconv.FormatFloat(seconds, 'f', -1, 64) + "s"

	if t.Ms >= minute {
		result = fmt.Sprintf("%dm %s", t.Ms/minute%secondsPerMinute, result)
	}
	if t.Ms >= hour {
		result = fmt.Sprintf("%dh %s", t.Ms/hour%hoursPerDay, result)
	}
	if t.Ms >= day {
		result = fmt.Sprintf("%dd %s", t.Ms/day, result)
	}
	return result
}

func (t *executiontime) formatDurationRoundrock() string {
	result := fmt.Sprintf("%dms", t.Ms%second)
	if t.Ms >= second {
		result = fmt.Sprintf("%ds %s", t.Ms/second%secondsPerMinute, result)
	}
	if t.Ms >= minute {
		result = fmt.Sprintf("%dm %s", t.Ms/minute%minutesPerHour, result)
	}
	if t.Ms >= hour {
		result = fmt.Sprintf("%dh %s", t.Ms/hour%hoursPerDay, result)
	}
	if t.Ms >= day {
		result = fmt.Sprintf("%dd %s", t.Ms/day, result)
	}
	return result
}

func (t *executiontime) formatDurationDallas() string {
	seconds := float64(t.Ms%minute) / second
	result := strconv.FormatFloat(seconds, 'f', -1, 64)

	if t.Ms >= minute {
		result = fmt.Sprintf("%d:%s", t.Ms/minute%minutesPerHour, result)
	}
	if t.Ms >= hour {
		result = fmt.Sprintf("%d:%s", t.Ms/hour%hoursPerDay, result)
	}
	if t.Ms >= day {
		result = fmt.Sprintf("%d:%s", t.Ms/day, result)
	}
	return result
}

func (t *executiontime) formatDurationGalveston() string {
	result := fmt.Sprintf("%02d:%02d:%02d", t.Ms/hour, t.Ms/minute%minutesPerHour, t.Ms%minute/second)
	return result
}

func (t *executiontime) formatDurationHouston() string {
	milliseconds := ".0"
	if t.Ms%second > 0 {
		// format milliseconds as a string with truncated trailing zeros
		milliseconds = strconv.FormatFloat(float64(t.Ms%second)/second, 'f', -1, 64)
		// at this point milliseconds looks like "0.5". remove the leading "0"
		if len(milliseconds) >= 1 {
			milliseconds = milliseconds[1:]
		}
	}

	result := fmt.Sprintf("%02d:%02d:%02d%s", t.Ms/hour, t.Ms/minute%minutesPerHour, t.Ms%minute/second, milliseconds)
	return result
}

func (t *executiontime) formatDurationAmarillo() string {
	// wholeNumber represents the value to the left of the decimal point (seconds)
	wholeNumber := t.Ms / second
	// decimalNumber represents the value to the right of the decimal point (milliseconds)
	decimalNumber := float64(t.Ms%second) / second

	// format wholeNumber as a string with thousands separators
	printer := message.NewPrinter(lang.English)
	result := printer.Sprintf("%d", wholeNumber)

	if decimalNumber > 0 {
		// format decimalNumber as a string with truncated trailing zeros
		decimalResult := strconv.FormatFloat(decimalNumber, 'f', -1, 64)
		// at this point decimalResult looks like "0.5"
		// remove the leading "0" and append
		if len(decimalResult) >= 1 {
			result += decimalResult[1:]
		}
	}
	result += "s"

	return result
}

func (t *executiontime) formatDurationRound() string {
	toRoundString := func(one, two int64, oneText, twoText string) string {
		if two == 0 {
			return fmt.Sprintf("%d%s", one, oneText)
		}
		return fmt.Sprintf("%d%s %d%s", one, oneText, two, twoText)
	}
	hours := t.Ms / hour % hoursPerDay
	if t.Ms >= day {
		return toRoundString(t.Ms/day, hours, "d", "h")
	}
	minutes := t.Ms / minute % secondsPerMinute
	if t.Ms >= hour {
		return toRoundString(hours, minutes, "h", "m")
	}
	seconds := (t.Ms % minute) / second
	if t.Ms >= minute {
		return toRoundString(minutes, seconds, "m", "s")
	}
	if t.Ms >= second {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dms", t.Ms%second)
}
