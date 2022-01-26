package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRust(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Rust 1.53.0", ExpectedString: "1.53.0", Version: "rustc 1.53.0 (4369396ce 2021-04-27)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "rustc",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.rs",
		}
		env, props := getMockedLanguageEnv(params)
		r := &Rust{}
		r.Init(props, env)
		assert.True(t, r.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
