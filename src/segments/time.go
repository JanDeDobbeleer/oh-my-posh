package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"time"
)

type Time struct {
	props properties.Properties
	env   platform.Environment

	CurrentDate time.Time
	Format      string
}

const (
	// TimeFormat uses the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format the current time
	TimeFormat properties.Property = "time_format"
)

func (t *Time) Template() string {
	return " {{ .CurrentDate | date .Format }} "
}

func (t *Time) Enabled() bool {
	// if no date set, use now(unit testing)
	t.Format = t.props.GetString(TimeFormat, "15:04:05")
	if t.CurrentDate.IsZero() {
		t.CurrentDate = time.Now()
	}
	return true
}

func (t *Time) Init(props properties.Properties, env platform.Environment) {
	t.props = props
	t.env = env
}
