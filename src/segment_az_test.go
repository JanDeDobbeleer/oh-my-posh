package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzSegment(t *testing.T) {
	cases := []struct {
		Case              string
		ExpectedEnabled   bool
		ExpectedString    string
		HasCLI            bool
		HasPowerShell     bool
		HasPowerShellUnix bool
		Template          string
	}{
		{
			Case:            "no config files found",
			ExpectedEnabled: false,
		},
		{
			Case:            "Az CLI Profile",
			ExpectedEnabled: true,
			ExpectedString:  "AzureCliCloud",
			Template:        "{{ .EnvironmentName }}",
			HasCLI:          true,
		},
		{
			Case:            "Az Pwsh Profile",
			ExpectedEnabled: true,
			ExpectedString:  "AzurePoshCloud",
			Template:        "{{ .EnvironmentName }}",
			HasPowerShell:   true,
		},
		{
			Case:              "Az Pwsh Profile",
			ExpectedEnabled:   true,
			ExpectedString:    "AzurePoshCloud",
			Template:          "{{ .EnvironmentName }}",
			HasPowerShellUnix: true,
		},
		{
			Case:            "Faulty template",
			ExpectedEnabled: true,
			ExpectedString:  incorrectTemplate,
			Template:        "{{ .Burp }}",
			HasPowerShell:   true,
		},
		{
			Case:            "PWSH",
			ExpectedEnabled: true,
			ExpectedString:  "PWSH",
			Template:        "{{ .Origin }}",
			HasPowerShell:   true,
		},
		{
			Case:            "CLI",
			ExpectedEnabled: true,
			ExpectedString:  "CLI",
			Template:        "{{ .Origin }}",
			HasCLI:          true,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		home := "/Users/posh"
		env.On("Home").Return(home)
		var azureProfile, azureRmContext, azureRMContext string
		if tc.HasCLI {
			content, _ := ioutil.ReadFile("./test/azureProfile.json")
			azureProfile = string(content)
		}
		if tc.HasPowerShell {
			content, _ := ioutil.ReadFile("./test/AzureRmContext.json")
			azureRmContext = string(content)
		}
		if tc.HasPowerShellUnix {
			content, _ := ioutil.ReadFile("./test/AzureRmContext.json")
			azureRMContext = string(content)
		}
		env.On("GOOS").Return(linuxPlatform)
		env.On("FileContent", filepath.Join(home, ".azure", "azureProfile.json")).Return(azureProfile)
		env.On("FileContent", filepath.Join(home, ".Azure", "AzureRmContext.json")).Return(azureRmContext)
		env.On("FileContent", filepath.Join(home, ".azure", "AzureRmContext.json")).Return(azureRMContext)
		az := &az{
			env:   env,
			props: properties{},
		}
		assert.Equal(t, tc.ExpectedEnabled, az.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, az), tc.Case)
	}
}
