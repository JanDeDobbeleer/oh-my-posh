package segments

import (
	"testing"
	"path"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestSitecoreSegment(t *testing.T) {
	cases := []struct {
		Case				string
		ExpectedString  	string
		ExpectedEnabled 	bool
		SitecoreFileExists 	bool
		UserFileExists 		bool
		UserFileContent		string
		DisplayDefault  	bool
	}{
		{ Case: "Disabled, no sitecore.json file and user.json file", ExpectedString: "", ExpectedEnabled: false, SitecoreFileExists: false, UserFileExists: false },
		{ Case: "Disabled, only sitecore.json file exists", ExpectedString: "", ExpectedEnabled: false, SitecoreFileExists: true, UserFileExists: false },
		{ Case: "Disabled, only user.json file exists", ExpectedString: "", ExpectedEnabled: false, SitecoreFileExists: false, UserFileExists: true },
		{ 
			Case: "Disabled, user.json is empty", 
			ExpectedString: "", 
			ExpectedEnabled: false, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: "",
		},
		{ 
			Case: "Disabled, user.json contains non-json text", 
			ExpectedString: "", 
			ExpectedEnabled: false, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: testUserJsonNotJsonFormat,
		},
		{ 
			Case: "Disabled with default endpoint", 
			ExpectedString: "default", 
			ExpectedEnabled: false, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: testUserJsonOnlyDefaultEnv,
			DisplayDefault: false,
		},
		{ 
			Case: "Enabled, user.json initial state", 
			ExpectedString: "default", 
			ExpectedEnabled: true, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: testUserJsonDefaultEmpty,
			DisplayDefault: true,
		},
		{ 
			Case: "Enabled, user.json with custom default endpoint and without endpoints", 
			ExpectedString: "MySuperEnv", 
			ExpectedEnabled: true, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: testUserJsonCustomDefaultEnvWithoutEndpoints,
		},
		{ 
			Case: "Enabled, user.json with custom default endpoint and configured endpoints", 
			ExpectedString: "myEnv (https://host.com)", 
			ExpectedEnabled: true, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: testUserJsonCustomDefaultEnv,
		},
		{ 
			Case: "Enabled, user.json with custom default endpoint and empty host", 
			ExpectedString: "envWithEmptyHost", 
			ExpectedEnabled: true, 
			SitecoreFileExists: true, 
			UserFileExists: true,
			UserFileContent: testUserJsonCustomDefaultEnvAndEmptyHost,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("HasFiles", "sitecore.json").Return(tc.SitecoreFileExists)
		env.On("HasFiles", path.Join(".sitecore", "user.json")).Return(tc.UserFileExists)
		env.On("FileContent", path.Join(".sitecore", "user.json")).Return(tc.UserFileContent)
		
		props := properties.Map{
			properties.DisplayDefault: tc.DisplayDefault,
		}

		sitecore := &Sitecore{}
		sitecore.Init(props, env)
		assert.Equal(t, tc.ExpectedEnabled, sitecore.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, sitecore.Template(), sitecore), tc.Case)
	}
}

var testUserJsonDefaultEmpty = `
{
	"endpoints": {}
}`

var testUserJsonCustomDefaultEnvWithoutEndpoints = `
{
	"endpoints": {},
	"defaultEndpoint": "MySuperEnv"
}`

var testUserJsonCustomDefaultEnv = `
{
	"endpoints": {
		"myEnv": {
			"host": "https://host.com"
		}
	},
	"defaultEndpoint": "myEnv"
}`

var testUserJsonCustomDefaultEnvAndEmptyHost = `
{
	"endpoints": {
		"myEnv": {
			"host": ""
		}
	},
	"defaultEndpoint": "envWithEmptyHost"
}`

var testUserJsonNotJsonFormat = `
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

var testUserJsonOnlyDefaultEnv = `
{
	"endpoints": {
		"default": {
			"host": "https://host.com"
		}
	}
}`