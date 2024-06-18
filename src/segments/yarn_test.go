package segments

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

func TestYarn(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "1.0.0", ExpectedString: "\U000F011B 1.0.0", Version: "1.0.0"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "yarn",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "package.json",
		}
		env, props := getMockedLanguageEnv(params)
		yarn := &Yarn{}
		yarn.Init(props, env)
		assert.True(t, yarn.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, yarn.Template(), yarn), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
