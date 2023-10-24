package segments

import (
	"io/fs"
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
			ExpectedEnabled:  true, // Segment should be enabled and visible
			Template:         "{{ .Version }}",
			ExpectedString:   "8.18.9",
		},
		{
			Case:             "Umbraco Folder and .csproj but NO web.config",
			HasUmbracoFolder: true,
			HasCsproj:        true,
			HasWebConfig:     false,
			ExpectedEnabled:  true, // Segment should be enabled and visible
			Template:         "{{ .Version }}",
			ExpectedString:   "12.1.2",
		},
		{
			Case:             "Umbraco Folder and .csproj with custom template",
			HasUmbracoFolder: true,
			HasCsproj:        true,
			ExpectedEnabled:  true,
			Template:         "Version:{{ .Version }} ModernUmbraco:{{ .IsModernUmbraco }} LegacyUmbraco:{{ .IsLegacyUmbraco }}",
			ExpectedString:   "Version:12.1.2 ModernUmbraco:true LegacyUmbraco:false",
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

		dirEntries := []fs.DirEntry{}
		if tc.HasUmbracoFolder {
			dirEntries = append(dirEntries, &MockDirEntry{
				name:  "Umbraco",
				isDir: true,
			})
		}

		if tc.HasCsproj {
			dirEntries = append(dirEntries, &MockDirEntry{
				name:  "MyProject.csproj",
				isDir: false,
			})
		}

		if tc.HasWebConfig {
			dirEntries = append(dirEntries, &MockDirEntry{
				name:  "web.config",
				isDir: false,
			})
		}

		env.On("LsDir", "/workspace/MyProject").Return(dirEntries)

		// Mocked these folder calls to return empty results
		// As the first test case will not find anything and then crawl up the folder tree
		env.On("LsDir", "/workspace").Return([]fs.DirEntry{})
		env.On("LsDir", "/").Return([]fs.DirEntry{})

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
