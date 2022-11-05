package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXMake(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "XMake 2.7.2-dev", ExpectedString: "2.7.2", Version: "xmake v2.7.2+dev.605b8e3e0"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "xmake",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "xmake.lua",
		}
		env, props := getMockedLanguageEnv(params)
		x := &XMake{}
		x.Init(props, env)
		assert.True(t, x.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, x.Template(), x), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
