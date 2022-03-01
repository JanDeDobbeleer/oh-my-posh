package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKotlin(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Kotlin 1.6.10", ExpectedString: "1.6.10", Version: "Kotlin version 1.6.10-release-923 (JRE 17.0.2+0)"},
		{Case: "Kotlin 1.6.0", ExpectedString: "1.6.0", Version: "Kotlin version 1.6.0-release-915 (JRE 17.0.2+0)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "kotlin",
			versionParam:  "-version",
			versionOutput: tc.Version,
			extension:     "*.kt",
		}
		env, props := getMockedLanguageEnv(params)
		k := &Kotlin{}
		k.Init(props, env)
		assert.True(t, k.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, k.Template(), k), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
