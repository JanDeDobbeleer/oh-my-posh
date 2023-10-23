package segments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestUmbracoSegment(t *testing.T) {
	cases := []struct {
		Case             string
		ExpectedEnabled  bool
		ExpectedString   string
		Template         string
		HasUmbracoFolder bool
		HasCsproj        bool
		HasWebConfig     bool
	}{
		{
			Case:             "No Umbraco folder found",
			HasUmbracoFolder: false,
			ExpectedEnabled:  false, // Segment should not be enabled
		},
		{
			Case:             "Umbraco Folder but NO web.config or .csproj",
			HasUmbracoFolder: true,
			HasCsproj:        false,
			HasWebConfig:     false,
			ExpectedEnabled:  false, // Segment should not be enabled
		},
		{
			Case:             "Umbraco Folder and web.config but NO .csproj",
			HasUmbracoFolder: true,
			HasCsproj:        false,
			HasWebConfig:     true,
			ExpectedEnabled:  true,     // Segment should be enabled and visible
			ExpectedString:   "8.15.0", // We are using the default template (by not specifying one) and expect just the version to be displayed
		},
		{
			Case:             "Umbraco Folder and .csproj but NO web.config",
			HasUmbracoFolder: true,
			HasCsproj:        true,
			HasWebConfig:     false,
			ExpectedEnabled:  true,     // Segment should be enabled and visible
			ExpectedString:   "12.1.2", // We are using the default template (by not specifying one) and expect just the version to be displayed
		},
		{
			Case:            "Umbraco Folder and .csproj with custom template",
			ExpectedEnabled: true,
			Template:        "Version:{{ .Version }} ModernUmbraco:{{ .IsModernUmbraco }} LegacyUmbraco:{{ .IsLegacyUmbraco }}",
			ExpectedString:  "Version:12.1.2 ModernUmbraco:true LegacyUmbraco:false",
		},
	}

	for _, tc := range cases {
		// Prepare/arrange the test
		env := new(mock.MockedEnvironment)
		var sampleCSProj, sampleWebConfig string

		if tc.HasCsproj {
			content, _ := os.ReadFile("../test/umbraco/MyProject.csproj")
			sampleCSProj = string(content)
		}
		if tc.HasWebConfig {
			content, _ := os.ReadFile("../test/umbraco/web.config")
			sampleWebConfig = string(content)
		}

		const umbracoProjectDirectory = "/workspace/MyProject"
		env.On("Pwd").Return(umbracoProjectDirectory)
		env.On("FileContent", filepath.Join(umbracoProjectDirectory, "MyProject.csproj")).Return(sampleCSProj)
		env.On("FileContent", filepath.Join(umbracoProjectDirectory, "web.config")).Return(sampleWebConfig)
		env.On("Debug", mock2.Anything)
		env.On("Trace", mock2.Anything, mock2.Anything)

		// TODO: HOW do I mock the folder/file structure so that Umbraco segmenet can test looping through parent folders
		// ******* Any help or pointers please Jan *******

		// Setup the Umbraco segment with the mocked environment & properties
		umb := &Umbraco{
			env: env,
		}

		// Assert the test results
		// Check if the segment should be enabled and
		// the rendered string matches what we expect when specifying a template for the segment
		assert.Equal(t, tc.ExpectedEnabled, umb.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, umb), tc.Case)
	}
}
