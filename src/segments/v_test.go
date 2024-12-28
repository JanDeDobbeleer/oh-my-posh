package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "V 0.4.9",
			ExpectedString: "0.4.9",
			Version:        "V 0.4.9 b487986",
		},
		{
			Case:           "V 0.4.8",
			ExpectedString: "0.4.8",
			Version:        "V 0.4.8 a123456",
		},
		{
			Case:           "V 0.4.7",
			ExpectedString: "0.4.7",
			Version:        "V 0.4.7 f789012",
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "v",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.v",
		}
		env, props := getMockedLanguageEnv(params)
		v := &V{}
		v.Init(props, env)
		assert.True(t, v.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, v.Template(), v), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
