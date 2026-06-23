package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestZvm(t *testing.T) {
	cases := []struct {
		Options         options.Map
		Case            string
		ListOutput      string
		ExpectedString  string
		HasCommand      bool
		ExpectedEnabled bool
	}{
		{
			Case:            "no zvm command",
			HasCommand:      false,
			ExpectedEnabled: false,
		},
		{
			Case:            "active version",
			HasCommand:      true,
			ListOutput:      "0.11.0\n[x]0.13.0\n0.12.0",
			ExpectedString:  "ZVM 0.13.0",
			ExpectedEnabled: true,
		},
		{
			Case:            "active version with spaced marker",
			HasCommand:      true,
			ListOutput:      "0.11.0\n[x] 0.13.0\n0.12.0",
			ExpectedString:  "ZVM 0.13.0",
			ExpectedEnabled: true,
		},
		{
			Case:            "custom icon",
			HasCommand:      true,
			ListOutput:      "0.11.0\n[x]0.13.0\n0.12.0",
			Options:         options.Map{ZigIcon: "⚡"},
			ExpectedString:  "⚡ 0.13.0",
			ExpectedEnabled: true,
		},
		{
			Case:            "no active version",
			HasCommand:      true,
			ListOutput:      "0.11.0\n0.12.0",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", "zvm").Return(tc.HasCommand)
		if tc.HasCommand {
			env.On("RunCommand", "zvm", []string{"--color=false", "list"}).Return(tc.ListOutput, nil)
		}

		props := tc.Options
		if props == nil {
			props = options.Map{}
		}

		zvm := &Zvm{}
		zvm.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, zvm.Enabled(), tc.Case)

		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, zvm.Template(), zvm), tc.Case)
		}

		env.AssertExpectations(t)
	}
}
