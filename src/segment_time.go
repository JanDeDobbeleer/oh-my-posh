package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"time"
)

type tempus struct {
	props properties.Properties
	env   environment.Environment

	CurrentDate time.Time
}

const (
	// TimeFormat uses the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format the current time
	TimeFormat properties.Property = "time_format"
)

func (t *tempus) template() string {
	return "{{ .CurrentDate | date \"" + t.props.GetString(TimeFormat, "15:04:05") + "\" }}"
}

func (t *tempus) enabled() bool {
	// if no date set, use now(unit testing)
	if t.CurrentDate.IsZero() {
		t.CurrentDate = time.Now()
	}
	return true
}

func (t *tempus) init(props properties.Properties, env environment.Environment) {
	t.props = props
	t.env = env
}
