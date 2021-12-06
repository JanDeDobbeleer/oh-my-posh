package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJulianDateSegment(t *testing.T) {
	cases := []struct {
		Case            string
		TestDate        time.Time
		ExpectedString  string
		ExpectedEnabled bool
		Template        string
		Error           error
	}{
		{
			Case:            "6th December 2021",
			TestDate:        time.Date(2021, time.Month(12), 6, 1, 10, 30, 0, time.UTC),
			Template:        "{{.Year}}{{.DayOfYear}}",
			ExpectedString:  "21340",
			ExpectedEnabled: true,
		},
		{
			Case:            "2nd January 2009 requires some padding",
			TestDate:        time.Date(2009, time.Month(1), 2, 1, 10, 30, 0, time.UTC),
			Template:        "{{.Year}}{{.DayOfYear}}",
			ExpectedString:  "09002",
			ExpectedEnabled: true,
		},
		{
			Case:            "20th March 2031",
			TestDate:        time.Date(2031, time.Month(3), 20, 1, 10, 30, 0, time.UTC),
			Template:        "{{.Year}}{{.DayOfYear}}",
			ExpectedString:  "31079",
			ExpectedEnabled: true,
		},
		{
			Case:            "Faulty template",
			TestDate:        time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC),
			Template:        "{{.DayOfYear}}{{.DoesntExist}}",
			ExpectedString:  incorrectTemplate,
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		var props properties = map[Property]interface{}{}
		env.On("getTimeNow").Return(tc.TestDate)

		if tc.Template != "" {
			props[SegmentTemplate] = tc.Template
		}

		jd := &juliandate{
			props: props,
			env:   env,
		}

		enabled := jd.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, jd.string(), tc.Case)
	}
}
