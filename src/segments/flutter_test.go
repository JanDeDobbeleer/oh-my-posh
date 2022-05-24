package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlutter(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Flutter 2.10.4", ExpectedString: "2.10.4", Version: "Flutter 2.10.4 • channel stable • https://github.com/flutter/flutter.git"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "flutter",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.dart",
		}
		env, props := getMockedLanguageEnv(params)
		d := &Flutter{}
		d.Init(props, env)
		assert.True(t, d.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, d.Template(), d), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
