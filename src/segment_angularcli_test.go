package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAngularCliVersionDisplayed(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Angular 12.2.9", ExpectedString: "12.2.9", Version: "Angular CLI: 12.2.9"},
	}

	for _, ta := range cases {
		params := &mockedLanguageParams{
			cmd:           "ng",
			versionParam:  "--version",
			versionOutput: ta.Version,
			extension:     "angular.json",
		}

		env, props := getMockedLanguageEnv(params)
		angularcli := &angularcli{}
		angularcli.init(props, env)
		assert.True(t, angularcli.enabled(), fmt.Sprintf("Failed in case: %s", ta.Case))
		assert.Equal(t, ta.ExpectedString, angularcli.string(), fmt.Sprintf("Failed in case: %s", ta.Case))
	}

}
