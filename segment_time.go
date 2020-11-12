package main

import (
	"time"
)

type tempus struct {
	props *properties
	env   environmentInfo
}

const (
	// TimeFormat uses the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format the current time
	TimeFormat Property = "time_format"
)

func (t *tempus) enabled() bool {
	return true
}

func (t *tempus) string() string {
	timeFormatProperty := t.props.getString(TimeFormat, "15:04:05")
	return time.Now().Format(timeFormatProperty)
}

func (t *tempus) init(props *properties, env environmentInfo) {
	t.props = props
	t.env = env
}
