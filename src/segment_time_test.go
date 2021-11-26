package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeSegmentTemplate(t *testing.T) {
	// set date for unit test
	currentDate := time.Now()
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		Template        string
	}{
		{
			Case:            "no template",
			Template:        "",
			ExpectedString:  currentDate.Format("15:04:05"),
			ExpectedEnabled: true,
		},
		{
			Case:            "time only",
			Template:        "{{.CurrentDate | date \"15:04:05\"}}",
			ExpectedString:  currentDate.Format("15:04:05"),
			ExpectedEnabled: true,
		},
		{
			Case:            "lowercase",
			Template:        "{{.CurrentDate | date \"January 02, 2006 15:04:05\" | lower }}",
			ExpectedString:  strings.ToLower(currentDate.Format("January 02, 2006 15:04:05")),
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		tempus := &tempus{
			env: env,
			props: map[Property]interface{}{
				SegmentTemplate: tc.Template,
			},
			CurrentDate: currentDate,
		}
		assert.Equal(t, tc.ExpectedEnabled, tempus.enabled())
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, tempus.string(), tc.Case)
		}
	}
}
