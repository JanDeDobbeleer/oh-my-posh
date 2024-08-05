package segments

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestUmbracoSegment(t *testing.T) {
	cases := []struct {
		Case                    string
		ExpectedString          string
		Template                string
		ExpectedEnabled         bool
		HasUmbracoFolder        bool
		HasCsproj               bool
		HasWebConfig            bool
		UseLegacyWebConfig      bool
		HasCSProjWithoutUmbraco bool
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
			Case:               "Umbraco Folder and web.config but NO .csproj and uses older web.config",
			HasUmbracoFolder:   true,
			HasCsproj:          false,
			HasWebConfig:       true,
			UseLegacyWebConfig: true,
			ExpectedEnabled:    true, // Segment should be enabled and visible
			Template:           "{{ .Version }}",
			ExpectedString:     "4.11.10",
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
			Case:                    "Umbraco Folder with a .csproj [without Umbraco] and a web.config",
			HasUmbracoFolder:        true,
			HasCsproj:               false,
			HasWebConfig:            true,
			HasCSProjWithoutUmbraco: true, // This is a .csproj file without Umbraco installed - so we can test it doesn't return but carries on checking for more csproj or web.config files
			ExpectedEnabled:         true, // Segment should not be visible
			Template:                "{{ .Version }}",
			ExpectedString:          "8.18.9",
		},
		{
			Case:                    "Umbraco Folder with a .csproj [without Umbraco] and NO web.config",
			HasUmbracoFolder:        true,
			HasCsproj:               false,
			HasWebConfig:            false,
			HasCSProjWithoutUmbraco: true,
			ExpectedEnabled:         false,
			Template:                "{{ .Version }}",
			ExpectedString:          "",
		},
		{
			Case:                    "Umbraco Folder with multiple .csproj's and NO web.config",
			HasUmbracoFolder:        true,
			HasCsproj:               true,
			HasWebConfig:            false,
			HasCSProjWithoutUmbraco: true,
			ExpectedEnabled:         true, // Segment should be enabled and visible
			Template:                "{{ .Version }}",
			ExpectedString:          "12.1.2",
		},
		{
			Case:             "Umbraco Folder and .csproj with custom template",
			HasUmbracoFolder: true,
			HasCsproj:        true,
			ExpectedEnabled:  true,
			Template:         "Version:{{ .Version }} ModernUmbraco:{{ .Modern }}",
			ExpectedString:   "Version:12.1.2 ModernUmbraco:true",
		},
	}

	for _, tc := range cases {
		// Prepare/arrange the test
		env := new(mock.Environment)
		var sampleCSProj, sampleWebConfig, sampleNonUmbracoCSProj string

		if tc.HasCsproj {
			content, _ := os.ReadFile("../test/umbraco/MyProject.csproj")
			sampleCSProj = string(content)
		}
		if tc.HasWebConfig {
			var filePath string
			if tc.UseLegacyWebConfig {
				filePath = "../test/umbraco/web.old.config"
			} else {
				filePath = "../test/umbraco/web.config"
			}

			content, _ := os.ReadFile(filePath)
			sampleWebConfig = string(content)
		}
		if tc.HasCSProjWithoutUmbraco {
			content, _ := os.ReadFile("../test/umbraco/ANonUmbracoProject.csproj")
			sampleNonUmbracoCSProj = string(content)
		}

		const umbracoProjectDirectory = "/workspace/MyProject"
		env.On("Pwd").Return(umbracoProjectDirectory)
		env.On("FileContent", filepath.Join(umbracoProjectDirectory, "MyProject.csproj")).Return(sampleCSProj)
		env.On("FileContent", filepath.Join(umbracoProjectDirectory, "ANonUmbracoProject.csproj")).Return(sampleNonUmbracoCSProj)
		env.On("FileContent", filepath.Join(umbracoProjectDirectory, "web.config")).Return(sampleWebConfig)
		env.On("Debug", testify_.Anything)

		if tc.HasUmbracoFolder {
			fileInfo := &runtime.FileInfo{
				Path:         "/workspace/MyProject/Umbraco",
				ParentFolder: "/workspace/MyProject",
				IsDir:        true,
			}

			env.On("HasParentFilePath", "umbraco", false).Return(fileInfo, nil)
		} else {
			env.On("HasParentFilePath", "Umbraco", false).Return(&runtime.FileInfo{}, errors.New("no such file or directory"))
			env.On("HasParentFilePath", "umbraco", false).Return(&runtime.FileInfo{}, errors.New("no such file or directory"))
		}

		dirEntries := []fs.DirEntry{}
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

		if tc.HasCSProjWithoutUmbraco {
			dirEntries = append(dirEntries, &MockDirEntry{
				name:  "ANonUmbracoProject.csproj",
				isDir: false,
			})
		}

		env.On("LsDir", umbracoProjectDirectory).Return(dirEntries)

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
