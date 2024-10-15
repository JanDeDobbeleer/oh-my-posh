package segments

import (
	"path"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestSitecoreSegment(t *testing.T) {
	cases := []struct {
		Case               string
		ExpectedString     string
		UserFileContent    string
		ExpectedEnabled    bool
		SitecoreFileExists bool
		UserFileExists     bool
		DisplayDefault     bool
	}{
		{Case: "Disabled, no sitecore.json file and user.json file", ExpectedString: "", ExpectedEnabled: false, SitecoreFileExists: false, UserFileExists: false},
		{Case: "Disabled, only sitecore.json file exists", ExpectedString: "", ExpectedEnabled: false, SitecoreFileExists: true, UserFileExists: false},
		{Case: "Disabled, only user.json file exists", ExpectedString: "", ExpectedEnabled: false, SitecoreFileExists: false, UserFileExists: true},
		{
			Case:               "Disabled, user.json is empty",
			ExpectedString:     "",
			ExpectedEnabled:    false,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    "",
		},
		{
			Case:               "Disabled, user.json contains non-json text",
			ExpectedString:     "",
			ExpectedEnabled:    false,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    testUserJSONNotJSONFormat,
		},
		{
			Case:               "Disabled with default endpoint",
			ExpectedString:     "default",
			ExpectedEnabled:    false,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    testUserJSONOnlyDefaultEnv,
			DisplayDefault:     false,
		},
		{
			Case:               "Enabled, user.json initial state",
			ExpectedString:     "default",
			ExpectedEnabled:    true,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    testUserJSONDefaultEmpty,
			DisplayDefault:     true,
		},
		{
			Case:               "Enabled, user.json with custom default endpoint and without endpoints",
			ExpectedString:     "MySuperEnv",
			ExpectedEnabled:    true,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    testUserJSONCustomDefaultEnvWithoutEndpoints,
		},
		{
			Case:               "Enabled, user.json with custom default endpoint and configured endpoints",
			ExpectedString:     "myEnv (https://host.com)",
			ExpectedEnabled:    true,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    testUserJSONCustomDefaultEnv,
		},
		{
			Case:               "Enabled, user.json with custom default endpoint and empty host",
			ExpectedString:     "envWithEmptyHost",
			ExpectedEnabled:    true,
			SitecoreFileExists: true,
			UserFileExists:     true,
			UserFileContent:    testUserJSONCustomDefaultEnvAndEmptyHost,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasFiles", "sitecore.json").Return(tc.SitecoreFileExists)
		env.On("HasFilesInDir", ".sitecore", "user.json").Return(tc.UserFileExists)
		env.On("FileContent", path.Join(".sitecore", "user.json")).Return(tc.UserFileContent)
		env.On("Debug", testify_.Anything)
		env.On("Error", testify_.Anything)

		props := properties.Map{
			properties.DisplayDefault: tc.DisplayDefault,
		}

		sitecore := &Sitecore{}
		sitecore.Init(props, env)
		assert.Equal(t, tc.ExpectedEnabled, sitecore.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, sitecore.Template(), sitecore), tc.Case)
	}
}

var testUserJSONDefaultEmpty = `
{
	"endpoints": {}
}`

var testUserJSONCustomDefaultEnvWithoutEndpoints = `
{
	"endpoints": {},
	"defaultEndpoint": "MySuperEnv"
}`

var testUserJSONCustomDefaultEnv = `
{
	"endpoints": {
		"myEnv": {
			"host": "https://host.com"
		}
	},
	"defaultEndpoint": "myEnv"
}`

var testUserJSONCustomDefaultEnvAndEmptyHost = `
{
	"endpoints": {
		"myEnv": {
			"host": ""
		}
	},
	"defaultEndpoint": "envWithEmptyHost"
}`

var testUserJSONNotJSONFormat = `
---
 doe: "a deer, a female deer"
 ray: "a drop of golden sun"
 pi: 3.14159
 xmas: true
 french-hens: 3
 calling-birds:
   - huey
   - dewey
   - louie
   - fred
 xmas-fifth-day:
   calling-birds: four
   french-hens: 3
   golden-rings: 5
   partridges:
     count: 1
     location: "a pear tree"
   turtle-doves: two`

var testUserJSONOnlyDefaultEnv = `
{
	"endpoints": {
		"default": {
			"host": "https://host.com"
		}
	}
}`
