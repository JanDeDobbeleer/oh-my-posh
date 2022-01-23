package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAngularCliVersionDisplayed(t *testing.T) {
	cases := []struct {
		Case        string
		FullVersion string
		Version     string
	}{
		{Case: "Angular 13.0.3", FullVersion: "13.0.3", Version: "{ \"name\": \"@angular/core\",\"version\": \"13.0.3\"}"},
		{Case: "Angular 11.0.1", FullVersion: "11.0.1", Version: "{ \"name\": \"@angular/core\",\"version\": \"11.0.1\"}"},
	}

	for _, ta := range cases {
		params := &mockedLanguageParams{
			extension: "angular.json",
		}

		var env = new(MockedEnvironment)
		// mock  getVersion methods
		env.On("Pwd").Return("/usr/home/dev/my-app")
		env.On("Home").Return("/usr/home")
		env.On("HasFiles", params.extension).Return(true)
		env.On("HasFilesInDir", "/usr/home/dev/my-app/node_modules/@angular/core", "package.json").Return(true)
		env.On("FileContent", "/usr/home/dev/my-app/node_modules/@angular/core/package.json").Return(ta.Version)
		env.On("TemplateCache").Return(&TemplateCache{
			Env: make(map[string]string),
		})
		props := properties{}
		angular := &angular{}
		angular.init(props, env)
		assert.True(t, angular.enabled(), fmt.Sprintf("Failed in case: %s", ta.Case))
		assert.Equal(t, ta.FullVersion, renderTemplate(env, angular.template(), angular), fmt.Sprintf("Failed in case: %s", ta.Case))
	}
}
