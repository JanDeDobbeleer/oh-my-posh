package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/alecthomas/assert"
)

func TestMojoTemplate(t *testing.T) {
	cases := []struct {
		Case            string
		Expected        string
		VirtualEnvName  string
		FetchVirtualEnv bool
		DisplayDefault  bool
		FetchVersion    bool
	}{
		{
			Case:            "Virtual environment is present",
			Expected:        "foo 24.5.0",
			VirtualEnvName:  "foo",
			FetchVirtualEnv: true,
			DisplayDefault:  true,
			FetchVersion:    true,
		},
		{
			Case:            "No virtual environment present",
			Expected:        "24.5.0",
			VirtualEnvName:  "",
			FetchVirtualEnv: true,
			DisplayDefault:  true,
			FetchVersion:    true,
		},
		{
			Case:            "Hide the virtual environment, but show the version",
			Expected:        "24.5.0",
			VirtualEnvName:  "foo",
			FetchVirtualEnv: false,
			DisplayDefault:  true,
			FetchVersion:    true,
		},
		{
			Case:            "Show the virtual environment, but hide the version",
			Expected:        "foo",
			VirtualEnvName:  "foo",
			FetchVirtualEnv: true,
			DisplayDefault:  true,
			FetchVersion:    false,
		},
		{
			Case:            "Show the default virtual environment",
			Expected:        "default 24.5.0",
			VirtualEnvName:  "default",
			FetchVirtualEnv: true,
			DisplayDefault:  true,
			FetchVersion:    true,
		},
		{
			Case:            "Hide the default virtual environment",
			Expected:        "24.5.0",
			VirtualEnvName:  "default",
			FetchVirtualEnv: true,
			DisplayDefault:  false,
			FetchVersion:    true,
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "mojo",
			versionParam:  "--version",
			versionOutput: "mojo 24.5.0 (e8aacb95)",
			extension:     "*.mojo",
		}
		env, props := getMockedLanguageEnv(params)
		env.On("Getenv", "PIXI_ENVIRONMENT_NAME").Return(tc.VirtualEnvName)
		props[properties.DisplayDefault] = tc.DisplayDefault
		props[properties.FetchVersion] = tc.FetchVersion
		props[FetchVirtualEnv] = tc.FetchVirtualEnv
		props[DisplayMode] = DisplayModeAlways

		mojo := &Mojo{}
		mojo.Init(props, env)
		assert.True(t, mojo.Enabled())
		assert.Equal(t, tc.Expected, renderTemplate(env, mojo.Template(), mojo), tc.Case)
	}
}
