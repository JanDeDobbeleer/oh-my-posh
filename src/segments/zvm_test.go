package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
)

func TestZvm(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		HasCommand     bool
		ColorState     string
		ListOutput     string
		Properties     properties.Map
		Template       string
		ExpectedIcon   string
	}{
		{
			Case:           "no zvm command",
			ExpectedString: "",
			HasCommand:     false,
			ColorState:     "",
			ListOutput:     "",
			Properties:     properties.Map{},
			Template:       " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
			ExpectedIcon:   DefaultZigIcon,
		},
		{
			Case:           "version with colors enabled",
			ExpectedString: "0.13.0",
			HasCommand:     true,
			ColorState:     "on",
			ListOutput:     "0.11.0\n[x]0.13.0\n0.12.0",
			Properties:     properties.Map{},
			Template:       " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
			ExpectedIcon:   DefaultZigIcon,
		},
		{
			Case:           "version with colors disabled",
			ExpectedString: "0.13.0",
			HasCommand:     true,
			ColorState:     "off",
			ListOutput:     "0.11.0\n[x]0.13.0\n0.12.0",
			Properties:     properties.Map{},
			Template:       " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
			ExpectedIcon:   DefaultZigIcon,
		},
		{
			Case:           "version with custom icon",
			ExpectedString: "0.13.0",
			HasCommand:     true,
			ColorState:     "on",
			ListOutput:     "0.11.0\n[x]0.13.0\n0.12.0",
			Properties: properties.Map{
				PropertyZigIcon: "⚡",
			},
			Template:     " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} ",
			ExpectedIcon: "⚡",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)

			// Mock HasCommand first
			env.On("HasCommand", "zvm").Return(tc.HasCommand)

			// Only set up other mocks if HasCommand is true
			if tc.HasCommand {
				// Mock color detection
				env.On("RunCommand", "zvm", []string{"--color"}).Return(tc.ColorState, nil)

				// Mock color state changes based on detected state
				if tc.ColorState == "on" {
					env.On("RunCommand", "zvm", []string{"--color", "false"}).Return("", nil)
					env.On("RunCommand", "zvm", []string{"--color", "true"}).Return("", nil)
				}

				// Mock version list command
				env.On("RunCommand", "zvm", []string{"list"}).Return(tc.ListOutput, nil)
			}

			zvm := &Zvm{}
			zvm.Init(tc.Properties, env)

			assert.Equal(t, tc.Template, zvm.Template())

			if tc.HasCommand {
				assert.True(t, zvm.Enabled())
				assert.Equal(t, tc.ExpectedString, zvm.Text())
				assert.Equal(t, tc.ExpectedIcon, zvm.ZigIcon)
			} else {
				assert.False(t, zvm.Enabled())
				assert.Empty(t, zvm.Text())
			}

			// Verify all expected calls were made
			env.AssertExpectations(t)
		})
	}
}

func TestColorStateDetection(t *testing.T) {
	cases := []struct {
		Case        string
		ColorOutput string
		Expected    colorState
	}{
		{
			Case:        "enabled - on",
			ColorOutput: "on",
			Expected:    colorState{enabled: true, valid: true},
		},
		{
			Case:        "enabled - yes",
			ColorOutput: "yes",
			Expected:    colorState{enabled: true, valid: true},
		},
		{
			Case:        "disabled - off",
			ColorOutput: "off",
			Expected:    colorState{enabled: false, valid: true},
		},
		{
			Case:        "disabled - no",
			ColorOutput: "no",
			Expected:    colorState{enabled: false, valid: true},
		},
		{
			Case:        "invalid state",
			ColorOutput: "invalid",
			Expected:    colorState{enabled: false, valid: false},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)
			env.On("RunCommand", "zvm", []string{"--color"}).Return(tc.ColorOutput, nil)

			cmd := &colorCommand{env: env}
			state := cmd.detectColorState()

			assert.Equal(t, tc.Expected.enabled, state.enabled)
			assert.Equal(t, tc.Expected.valid, state.valid)
			env.AssertExpectations(t)
		})
	}
}
