package segments

import (
	"oh-my-posh/mock"
	"oh-my-posh/properties"
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
		env := new(mock.MockedEnvironment)
		tempus := &Time{
			env:         env,
			props:       properties.Map{},
			CurrentDate: currentDate,
		}
		assert.Equal(t, tc.ExpectedEnabled, tempus.Enabled())
		if tc.Template == "" {
			tc.Template = tempus.Template()
		}
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, tempus), tc.Case)
		}
	}
}
