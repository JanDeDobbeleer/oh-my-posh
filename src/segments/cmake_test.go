package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmake(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Cmake 3.23.2", ExpectedString: "3.23.2", Version: "cmake version 3.23.2"},
		{Case: "Cmake 2.3.13", ExpectedString: "2.3.12", Version: "cmake version 2.3.12"},
		{Case: "", ExpectedString: "", Version: ""},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "cmake",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.cmake",
		}
		env, props := getMockedLanguageEnv(params)
		c := &Cmake{}
		c.Init(props, env)
		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, c.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.Template(), c), failMsg)
	}
}
