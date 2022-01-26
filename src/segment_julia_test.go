package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJulia(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Julia 1.6.0", ExpectedString: "1.6.0", Version: "julia version 1.6.0"},
		{Case: "Julia 1.6.1", ExpectedString: "1.6.1", Version: "julia version 1.6.1"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "julia",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.jl",
		}
		env, props := getMockedLanguageEnv(params)
		j := &Julia{}
		j.init(props, env)
		assert.True(t, j.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, j.template(), j), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
