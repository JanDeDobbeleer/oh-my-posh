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
		Template       string
	}{
		{Case: "bazel 4.0.0", ExpectedString: "<LINK>https://docs.bazel.build/versions/4.0.0<TEXT>\ue63a</TEXT></LINK> 4.0.0", Version: "bazel 4.0.0", Template: ""},
		{Case: "bazel 5.4.1", ExpectedString: "<LINK>https://docs.bazel.build/versions/5.4.1<TEXT>\ue63a</TEXT></LINK> 5.4.1", Version: "bazel 5.4.1", Template: ""},
		{Case: "bazel 6.4.0", ExpectedString: "<LINK>https://bazel.build/versions/6.4.0<TEXT>\ue63a</TEXT></LINK> 6.4.0", Version: "bazel 6.4.0", Template: ""},
		{Case: "bazel 7.1.1", ExpectedString: "<LINK>https://bazel.build/versions/7.1.1<TEXT>\ue63a</TEXT></LINK> 7.1.1", Version: "bazel 7.1.1", Template: ""},
		{Case: "bazel 10.11.12", ExpectedString: "<LINK>https://bazel.build/versions/10.11.12<TEXT>\ue63a</TEXT></LINK> 10.11.12", Version: "bazel 10.11.12", Template: ""},
		{Case: "", ExpectedString: "\ue63a err parsing info from bazel with", Version: "", Template: ""},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "bazel",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.bazel",
		}
		env, props := getMockedLanguageEnv(params)
		props[Icon] = "\ue63a"
		b := &Bazel{}
		b.Init(props, env)
		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, b.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, b.Template(), b), failMsg)
	}
}
