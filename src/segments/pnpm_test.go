package segments

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

func TestPnpm(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "1.0.0", ExpectedString: "\U000F02C1 1.0.0", Version: "1.0.0"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "pnpm",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "package.json",
		}
		env, props := getMockedLanguageEnv(params)
		pnpm := &Pnpm{}
		pnpm.Init(props, env)
		assert.True(t, pnpm.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, pnpm.Template(), pnpm), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
