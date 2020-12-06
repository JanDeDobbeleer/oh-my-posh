package main

import (
	"time"
)

type executiontime struct {
	props  *properties
	env    environmentInfo
	output string
}

const (
	// ThresholdProperty represents minimum duration (milliseconds) required to enable this segment
	ThresholdProperty Property = "threshold"
)

func (t *executiontime) enabled() bool {
	executionTimeMs := t.env.executionTime()
	thresholdMs := t.props.getFloat64(ThresholdProperty, float64(500))
	if executionTimeMs < thresholdMs {
		return false
	}

	duration := time.Duration(executionTimeMs) * time.Millisecond
	t.output = duration.String()
	return t.output != ""
}

func (t *executiontime) string() string {
	return t.output
}

func (t *executiontime) init(props *properties, env environmentInfo) {
	t.props = props
	t.env = env
}
