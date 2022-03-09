package segments

import (
	"io/ioutil"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"oh-my-posh/template"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		HasCLI          bool
		HasPowerShell   bool
		Template        string
		Source          string
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
			Case:            "Az Pwsh Profile",
			ExpectedEnabled: true,
			ExpectedString:  "AzurePoshCloud",
			Template:        "{{ .EnvironmentName }}",
			HasPowerShell:   true,
		},
		{
			Case:            "Faulty template",
			ExpectedEnabled: true,
			ExpectedString:  template.IncorrectTemplate,
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
		{
			Case:            "Az CLI Profile only",
			ExpectedEnabled: true,
			ExpectedString:  "AzureCliCloud",
			Template:        "{{ .EnvironmentName }}",
			HasCLI:          true,
			Source:          cli,
		},
		{
			Case:            "Az CLI Profile only - disabled",
			ExpectedEnabled: false,
			Template:        "{{ .EnvironmentName }}",
			HasCLI:          false,
			Source:          cli,
		},
		{
			Case:            "PowerShell Profile only",
			ExpectedEnabled: true,
			ExpectedString:  "AzurePoshCloud",
			Template:        "{{ .EnvironmentName }}",
			HasPowerShell:   true,
			Source:          pwsh,
		},
		{
			Case:            "Az CLI Profile only - disabled",
			ExpectedEnabled: false,
			Template:        "{{ .EnvironmentName }}",
			Source:          pwsh,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		home := "/Users/posh"
		env.On("Home").Return(home)
		var azureProfile, azureRmContext string
		if tc.HasCLI {
			content, _ := ioutil.ReadFile("../test/azureProfile.json")
			azureProfile = string(content)
		}
		if tc.HasPowerShell {
			content, _ := ioutil.ReadFile("../test/AzureRmContext.json")
			azureRmContext = string(content)
		}
		env.On("GOOS").Return(environment.LinuxPlatform)
		env.On("FileContent", filepath.Join(home, ".azure", "azureProfile.json")).Return(azureProfile)
		env.On("FileContent", filepath.Join(home, ".azure", "AzureRmContext.json")).Return(azureRmContext)
		env.On("Getenv", "AZURE_CONFIG_DIR").Return("")
		if tc.Source == "" {
			tc.Source = firstMatch
		}
		az := &Az{
			env: env,
			props: properties.Map{
				Source: tc.Source,
			},
		}
		assert.Equal(t, tc.ExpectedEnabled, az.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, az), tc.Case)
	}
}
