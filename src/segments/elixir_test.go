package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElixir(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "elixir 1.14.2",
			ExpectedString: "1.14.2",
			Version:        "Erlang/OTP 25 [erts-13.1.3] [source] [64-bit] [smp:8:8] [ds:8:8:10] [async-threads:1] [jit] [dtrace]\n\nElixir 1.14.2 (compiled with Erlang/OTP 25)",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "elixir",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.ex",
		}
		env, props := getMockedLanguageEnv(params)
		r := &Elixir{}
		r.Init(props, env)
		assert.True(t, r.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
