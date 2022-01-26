package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDart(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Dart 2.12.4", ExpectedString: "2.12.4", Version: "Dart SDK version: 2.12.4 (stable) (Thu Apr 15 12:26:53 2021 +0200) on \"macos_x64\""},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "dart",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.dart",
		}
		env, props := getMockedLanguageEnv(params)
		d := &Dart{}
		d.Init(props, env)
		assert.True(t, d.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, d.Template(), d), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
