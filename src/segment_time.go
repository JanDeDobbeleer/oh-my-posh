package main

import "time"

type tempus struct {
	props Properties
	env   Environment

	CurrentDate time.Time
}

const (
	// TimeFormat uses the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format the current time
	TimeFormat Property = "time_format"
)

func (t *tempus) template() string {
	return "{{ .CurrentDate | date \"" + t.props.getString(TimeFormat, "15:04:05") + "\" }}"
}

func (t *tempus) enabled() bool {
	// if no date set, use now(unit testing)
	if t.CurrentDate.IsZero() {
		t.CurrentDate = time.Now()
	}
	return true
}

func (t *tempus) init(props Properties, env Environment) {
	t.props = props
	t.env = env
}
