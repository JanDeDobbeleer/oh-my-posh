package segments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestAzSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		Template        string
		Source          string
		ExpectedEnabled bool
		HasCLI          bool
		HasPowerShell   bool
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
			ExpectedString:  "<.Data.Burp>: can't evaluate field Burp in type template.Data",
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
			Source:          Cli,
		},
		{
			Case:            "Az CLI Profile only - disabled",
			ExpectedEnabled: false,
			Template:        "{{ .EnvironmentName }}",
			HasCLI:          false,
			Source:          Cli,
		},
		{
			Case:            "PowerShell Profile only",
			ExpectedEnabled: true,
			ExpectedString:  "AzurePoshCloud",
			Template:        "{{ .EnvironmentName }}",
			HasPowerShell:   true,
			Source:          Pwsh,
		},
		{
			Case:            "Az CLI Profile only - disabled",
			ExpectedEnabled: false,
			Template:        "{{ .EnvironmentName }}",
			Source:          Pwsh,
		},
		{
			Case:            "Az CLI account type",
			ExpectedEnabled: true,
			ExpectedString:  "user",
			Template:        "{{ .User.Type }}",
			HasCLI:          true,
			Source:          Cli,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(poshHome)
		env.On("Flags").Return(&runtime.Flags{})

		var azureProfile, azureRmContext string

		if tc.HasCLI {
			content, _ := os.ReadFile("../test/azureProfile.json")
			azureProfile = string(content)
		}
		if tc.HasPowerShell {
			content, _ := os.ReadFile("../test/AzureRmContext.json")
			azureRmContext = string(content)
		}

		env.On("GOOS").Return(runtime.LINUX)
		env.On("FileContent", filepath.Join(poshHome, ".azure", "azureProfile.json")).Return(azureProfile)
		env.On("Getenv", "POSH_AZURE_SUBSCRIPTION").Return(azureRmContext)
		env.On("Getenv", "AZURE_CONFIG_DIR").Return("")

		if tc.HasCLI {
			env.On("HasFilesInDir", filepath.Clean("/Users/posh/.azure"), "azureProfile.json").Return(true)
		} else {
			env.On("HasFilesInDir", filepath.Clean("/Users/posh/.azure"), "azureProfile.json").Return(false)
			env.On("HasFilesInDir", filepath.Clean("/Users/posh/.Azure"), "azureProfile.json").Return(false)
		}

		if tc.Source == "" {
			tc.Source = FirstMatch
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
