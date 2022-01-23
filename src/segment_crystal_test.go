package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrystal(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Crystal 1.0.0", ExpectedString: "1.0.0", Version: "Crystal 1.0.0 (2021-03-22)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "crystal",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.cr",
		}
		env, props := getMockedLanguageEnv(params)
		c := &crystal{}
		c.init(props, env)
		assert.True(t, c.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, c.template(), c), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
