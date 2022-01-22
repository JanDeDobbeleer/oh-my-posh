package main

import "time"

type tempus struct {
	props        Properties
	env          Environment
	templateText string
	CurrentDate  time.Time
}

const (
	// TimeFormat uses the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format the current time
	TimeFormat Property = "time_format"
)

func (t *tempus) enabled() bool {
	// if no date set, use now(unit testing)
	if t.CurrentDate.IsZero() {
		t.CurrentDate = time.Now()
	}

	return true
}

func (t *tempus) string() string {
	segmentTemplate := t.props.getString(SegmentTemplate, "")
	if segmentTemplate != "" {
		template := &textTemplate{
			Template: segmentTemplate,
			Context:  t,
			Env:      t.env,
		}
		var err error
		t.templateText, err = template.render()
		if err != nil {
			t.templateText = err.Error()
		}
		if len(t.templateText) > 0 {
			return t.templateText
		}
	}
	// keep old behaviour if no template for now
	timeFormatProperty := t.props.getString(TimeFormat, "15:04:05")
	return t.CurrentDate.Format(timeFormatProperty)
}

func (t *tempus) init(props Properties, env Environment) {
	t.props = props
	t.env = env
}
