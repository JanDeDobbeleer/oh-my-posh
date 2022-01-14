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
		{Case: "Angular 13.0.3", ExpectedString: "13.0.3", Version: "{ \"name\": \"@angular/core\",\"version\": \"13.0.3\"}"},
		{Case: "Angular 11.0.1", ExpectedString: "11.0.1", Version: "{ \"name\": \"@angular/core\",\"version\": \"11.0.1\"}"},
	}

	for _, ta := range cases {
		params := &mockedLanguageParams{
			extension: "angular.json",
		}

		var env = new(MockedEnvironment)
		// mock  getVersion methods
		env.On("getcwd", nil).Return("/usr/home/dev/my-app")
		env.On("homeDir", nil).Return("/usr/home")
		env.On("hasFiles", params.extension).Return(true)
		env.On("hasFilesInDir", "/usr/home/dev/my-app/node_modules/@angular/core", "package.json").Return(true)
		env.On("getFileContent", "/usr/home/dev/my-app/node_modules/@angular/core/package.json").Return(ta.Version)
		env.onTemplate()
		props := properties{}
		angular := &angular{}
		angular.init(props, env)
		assert.True(t, angular.enabled(), fmt.Sprintf("Failed in case: %s", ta.Case))
		assert.Equal(t, ta.ExpectedString, angular.string(), fmt.Sprintf("Failed in case: %s", ta.Case))
	}
}
