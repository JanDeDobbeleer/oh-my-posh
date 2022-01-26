package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhp(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "PHP 6.1.0", ExpectedString: "6.1.0", Version: "PHP 6.1.0(cli) (built: Jul  2 2021 03:59:48) ( NTS )"},
		{Case: "php 7.4.21", ExpectedString: "7.4.21", Version: "PHP 7.4.21 (cli) (built: Jul  2 2021 03:59:48) ( NTS )"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "php",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.php",
		}
		env, props := getMockedLanguageEnv(params)
		j := &Php{}
		j.Init(props, env)
		assert.True(t, j.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, j.Template(), j), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
