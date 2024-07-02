package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/stretchr/testify/assert"
)

func TestElixir(t *testing.T) {
	cases := []struct {
		Case                string
		ExpectedString      string
		ElixirVersionOutput string
		AsdfVersionOutput   string
		HasAsdf             bool
		AsdfExitCode        int
	}{
		{
			Case:                "Version without asdf",
			ExpectedString:      "1.14.2",
			ElixirVersionOutput: "Erlang/OTP 25 [erts-13.1.3] [source] [64-bit] [smp:8:8] [ds:8:8:10] [async-threads:1] [jit] [dtrace]\n\nElixir 1.14.2 (compiled with Erlang/OTP 25)",
		},
		{
			Case:                "Version with asdf",
			ExpectedString:      "1.14.2",
			HasAsdf:             true,
			AsdfVersionOutput:   "elixir          1.14.2-otp-25   /path/to/.tool-versions",
			ElixirVersionOutput: "Should not be used",
		},
		{
			Case:                "Version with asdf not set: should fall back to elixir --version",
			ExpectedString:      "1.14.2",
			HasAsdf:             true,
			AsdfVersionOutput:   "elixir             ______          No version is set. Run \"asdf <global|shell|local> elixir <version>\"",
			AsdfExitCode:        126,
			ElixirVersionOutput: "Erlang/OTP 25 [erts-13.1.3] [source] [64-bit] [smp:8:8] [ds:8:8:10] [async-threads:1] [jit] [dtrace]\n\nElixir 1.14.2 (compiled with Erlang/OTP 25)",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "elixir",
			versionParam:  "--version",
			versionOutput: tc.ElixirVersionOutput,
			extension:     "*.ex",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("HasCommand", "asdf").Return(tc.HasAsdf)
		var asdfErr error
		if tc.AsdfExitCode != 0 {
			asdfErr = &runtime.CommandError{ExitCode: tc.AsdfExitCode}
		}
		env.On("RunCommand", "asdf", []string{"current", "elixir"}).Return(tc.AsdfVersionOutput, asdfErr)

		r := &Elixir{}
		r.Init(props, env)
		assert.True(t, r.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
