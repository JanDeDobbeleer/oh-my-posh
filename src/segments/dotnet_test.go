package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/constants"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/stretchr/testify/assert"
)

func TestDotnetSegment(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		ExitCode int
		Version  string
	}{
		{Case: "Unsupported version", Expected: "\uf071", ExitCode: constants.DotnetExitCode, Version: "3.1.402"},
		{Case: "Regular version", Expected: "3.1.402", Version: "3.1.402"},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "dotnet",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.cs",
		}
		env, props := getMockedLanguageEnv(params)

		if tc.ExitCode != 0 {
			env.Unset("RunCommand")
			err := &platform.CommandError{ExitCode: tc.ExitCode}
			env.On("RunCommand", "dotnet", []string{"--version"}).Return("", err)
		}

		dotnet := &Dotnet{}
		dotnet.Init(props, env)

		assert.True(t, dotnet.Enabled())
		assert.Equal(t, tc.Expected, renderTemplate(env, dotnet.Template(), dotnet), tc.Case)
	}
}
