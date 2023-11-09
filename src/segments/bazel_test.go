package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const icon = "\ue63a"

func TestBazel(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		Template       string
	}{
		{Case: "bazel 6.4.0", ExpectedString: fmt.Sprintf("«%s»(https://bazel.build/versions/6.4.0) 6.4.0", icon), Version: "bazel 6.4.0", Template: ""},
		{Case: "bazel 10.11.12", ExpectedString: fmt.Sprintf("«%s»(https://docs.bazel.build/versions/3.7.0) 3.7.0", icon), Version: "bazel 3.7.0"},
		{Case: "", ExpectedString: fmt.Sprintf("%s err parsing info from bazel with", icon), Version: ""},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "bazel",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.bazel",
		}
		env, props := getMockedLanguageEnv(params)
		props[Icon] = icon
		b := &Bazel{}
		b.Init(props, env)
		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, b.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), failMsg)
	}
}
