package segments

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

func TestNpm(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "1.0.0", ExpectedString: "\ue71e 1.0.0", Version: "1.0.0"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "npm",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "package.json",
		}
		env, props := getMockedLanguageEnv(params)
		npm := &Npm{}
		npm.Init(props, env)
		assert.True(t, npm.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, npm.Template(), npm), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
