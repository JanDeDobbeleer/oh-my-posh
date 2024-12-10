package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/stretchr/testify/assert"
)

func TestZvm(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		HasCommand     bool
		ListOutput     string
		Properties     properties.Map
		Template       string
	}{
		{
			Case:           "no zvm command",
			ExpectedString: "",
			HasCommand:     false,
			ListOutput:     "",
			Properties:     properties.Map{},
			Template:       " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
		},
		{
			Case:           "version 0.13.0 active",
			ExpectedString: "0.13.0",
			HasCommand:     true,
			ListOutput:     "0.11.0\n\x1b[32m0.13.0\x1b[0m\n0.12.0",
			Properties:     properties.Map{},
			Template:       " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
		},
		{
			Case:           "version 0.13.0 active with icon",
			ExpectedString: "0.13.0",
			HasCommand:     true,
			ListOutput:     "0.11.0\n\x1b[32m0.13.0\x1b[0m\n0.12.0",
			Properties: properties.Map{
				ZvmIcon: "⚡",
			},
			Template: " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
		},
		{
			Case:           "version 0.12.0-dev active",
			ExpectedString: "0.12.0-dev.1234+abcdef",
			HasCommand:     true,
			ListOutput:     "0.11.0\n\x1b[32m0.12.0-dev.1234+abcdef\x1b[0m\n0.12.0",
			Properties:     properties.Map{},
			// Change all test cases to expect the actual template
			Template: " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("HasCommand", "zvm").Return(tc.HasCommand)
			env.On("RunCommand", "zvm", []string{"list"}).Return(tc.ListOutput, nil)

			zvm := &Zvm{}
			zvm.Init(tc.Properties, env)

			assert.Equal(t, tc.Template, zvm.Template())

			if tc.HasCommand {
				assert.True(t, zvm.Enabled())
				assert.Equal(t, tc.ExpectedString, zvm.Version)
				if icon, ok := tc.Properties[ZvmIcon]; ok {
					assert.Equal(t, icon, zvm.ZigIcon)
				}
			} else {
				assert.False(t, zvm.Enabled())
			}
		})
	}
}
