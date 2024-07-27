package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestStatusWriterEnabled(t *testing.T) {
	cases := []struct {
		Status   int
		Expected bool
		Template string
	}{
		{Status: 102, Expected: true},
		{Status: 0, Expected: false},
		{Status: -1, Expected: true},
		{Status: 144, Expected: true, Template: "{{}}"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("StatusCodes").Return(tc.Status, "")
		env.On("TemplateCache").Return(&cache.Template{
			Code: 133,
		})
		env.On("Error", testify_.Anything).Return(nil)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("Flags").Return(&runtime.Flags{})

		props := properties.Map{}
		if len(tc.Template) > 0 {
			props[StatusTemplate] = tc.Template
		}

		s := &Status{}
		s.Init(props, env)

		assert.Equal(t, tc.Expected, s.Enabled())
	}
}

func TestFormatStatus(t *testing.T) {
	cases := []struct {
		Case       string
		Status     int
		PipeStatus string
		Template   string
		Separator  string
		Expected   string
	}{
		{
			Case:      "No PipeStatus",
			Status:    12,
			Template:  "{{ .Code }}",
			Separator: "|",
			Expected:  "12",
		},
		{
			Case:       "Defaults",
			PipeStatus: "0 127 0",
			Template:   "{{ .Code }}",
			Separator:  "|",
			Expected:   "0|127|0",
		},
		{
			Case:       "No integer",
			PipeStatus: "0 err 0",
			Template:   "{{ .Code }}",
			Separator:  "|",
			Expected:   "0|err|0",
		},
		{
			Case:       "Incorrect template",
			PipeStatus: "1 0 0",
			Template:   "{{}}",
			Separator:  "|",
			Expected:   "1|0|0",
		},
		{
			Case:       "Advanced template",
			PipeStatus: "1 0 0",
			Template:   "{{ if eq .Code 0 }}\uf058{{ else }}\uf071{{ end }}",
			Separator:  "|",
			Expected:   "\uf071|\uf058|\uf058",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("TemplateCache").Return(&cache.Template{
			Code: 133,
		})
		env.On("Error", testify_.Anything).Return(nil)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("Flags").Return(&runtime.Flags{})

		props := properties.Map{
			StatusTemplate:  tc.Template,
			StatusSeparator: tc.Separator,
		}

		s := &Status{}
		s.Init(props, env)

		assert.Equal(t, tc.Expected, s.formatStatus(tc.Status, tc.PipeStatus), tc.Case)
	}
}
