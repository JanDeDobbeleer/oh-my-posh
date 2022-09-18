package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeno(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Deno 1.25.2", ExpectedString: "1.25.2", Version: "deno 1.25.2 (release, aarch64-apple-darwin)\nv8 10.6.194.5\ntypescript 4.7.4"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "deno",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.js",
		}
		env, props := getMockedLanguageEnv(params)
		d := &Deno{}
		d.Init(props, env)
		assert.True(t, d.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, d.Template(), d), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
