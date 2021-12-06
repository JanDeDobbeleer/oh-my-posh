package main

import (
	"fmt"
)

type juliandate struct {
	props properties
	env   environmentInfo

	DayOfYear string
	Year      string
}

func (j *juliandate) enabled() bool {
	now := j.env.getTimeNow()
	// returns the last two digits of the year, for example year 2021 will return 21, and 2009 returns 09
	j.Year = fmt.Sprintf("%02d", now.Year()%100)
	// returns the day of the year, in the range [001,365] for non-leap years, and [001,366] in leap years.
	j.DayOfYear = fmt.Sprintf("%03d", now.YearDay())

	return true
}

func (j *juliandate) string() string {
	segmentTemplate := j.props.getString(SegmentTemplate, "{{.Year}}{{.DayOfYear}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  j,
		Env:      j.env,
	}

	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (j *juliandate) init(props properties, env environmentInfo) {
	j.props = props
	j.env = env
}
