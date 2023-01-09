package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVala(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "vala 0.48.17",
			ExpectedString: "0.48.17",
			Version:        "Vala 0.48.17",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "vala",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.vala",
		}
		env, props := getMockedLanguageEnv(params)
		v := &Vala{}
		v.Init(props, env)
		assert.True(t, v.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, v.Template(), v), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
