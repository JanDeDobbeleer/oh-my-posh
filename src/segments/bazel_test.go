package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBazel(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "bazel 6.4.0", ExpectedString: "6.4.0", Version: "bazel 6.4.0"},
		{Case: "bazel 10.11.12", ExpectedString: "10.11.12", Version: "bazel 10.11.12"},
		{Case: "bazel error", ExpectedString: "err parsing info from bazel with", Version: ""},
		{Case: "", ExpectedString: "err parsing info from bazel with", Version: ""},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "bazel",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.bazel",
		}
		env, props := getMockedLanguageEnv(params)
		b := &Bazel{}
		b.Init(props, env)
		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, b.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), failMsg)
	}
}
