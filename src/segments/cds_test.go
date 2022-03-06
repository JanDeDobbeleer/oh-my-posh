package segments

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCdsSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		File            string
		Template        string
		Version         string
		PackageJSON     string
		DisplayMode     string
	}{
		{
			Case:            "1) cds 5.5.0 - file .cdsrc.json present",
			ExpectedString:  "5.5.0",
			ExpectedEnabled: true,
			File:            ".cdsrc.json",
			Version:         "@sap/cds: 5.5.0\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "2) cds 5.5.1 - file some.cds",
			ExpectedString:  "5.5.1",
			ExpectedEnabled: true,
			File:            "some.cds",
			Version:         "@sap/cds: 5.5.1\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "3) cds 5.5.2 - no files",
			ExpectedString:  "",
			ExpectedEnabled: false,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "4) cds 5.5.3 - package.json dependency",
			ExpectedString:  "5.5.3",
			ExpectedEnabled: true,
			Version:         "@sap/cds: 5.5.3\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:     "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/cds\": \"^5\" } }",
			DisplayMode:     DisplayModeContext,
		},
		{
			Case:            "4) cds 5.5.4 - package.json dependency, major + minor",
			ExpectedString:  "5.5",
			ExpectedEnabled: true,
			Template:        "{{ .Major }}.{{ .Minor }}",
			Version:         "@sap/cds: 5.5.4\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:     "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/cds\": \"^5\" } }",
			DisplayMode:     DisplayModeContext,
		},
		{
			Case:            "5) cds 5.5.5 - package.json present, no dependency, no files",
			ExpectedString:  "",
			ExpectedEnabled: false,
			Version:         "@sap/cds: 5.5.5\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:     "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/some\": \"^5\" } }",
			DisplayMode:     DisplayModeContext,
		},
		{
			Case:            "6) cds 5.5.9 - display always",
			ExpectedString:  "5.5.9",
			ExpectedEnabled: true,
			Version:         "@sap/cds: 5.5.9\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:     "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/cds\": \"^5\" } }",
			DisplayMode:     DisplayModeAlways,
		},
		{
			Case:            "7) cds 5.5.9 - package.json, no dependencies section",
			ExpectedString:  "",
			ExpectedEnabled: false,
			Version:         "@sap/cds: 5.5.9\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:     "{ \"name\": \"my-app\" }",
			DisplayMode:     DisplayModeContext,
		},
		{
			Case:            "8) cds 5.5.0 - file .cdsrc-private.json present",
			ExpectedString:  "5.5.0",
			ExpectedEnabled: true,
			File:            ".cdsrc-private.json",
			Version:         "@sap/cds: 5.5.0\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			DisplayMode:     DisplayModeFiles,
		},
	}

	for _, tc := range cases {
		var env = new(mock.MockedEnvironment)
		env.On("HasCommand", "cds").Return(true)
		env.On("RunCommand", "cds", []string{"--version"}).Return(tc.Version, nil)
		env.On("Pwd").Return("/usr/home/dev/my-app")
		env.On("Home").Return("/usr/home")

		if tc.PackageJSON != "" {
			env.On("HasFiles", "package.json").Return(true)
			env.On("FileContent", "package.json").Return(tc.PackageJSON)
		} else {
			env.On("HasFiles", "package.json").Return(false)
		}

		cds := &Cds{}

		props := properties.Map{
			"display_mode": tc.DisplayMode,
		}

		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})

		if tc.Template == "" {
			tc.Template = cds.Template()
		}

		if tc.DisplayMode == "" {
			tc.DisplayMode = DisplayModeContext
		}

		cds.Init(props, env)

		for _, f := range cds.language.extensions {
			match, err := filepath.Match(f, tc.File)

			if err != nil {
				t.Fail()
			}

			env.On("HasFiles", f).Return(match)
		}

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.Equal(t, tc.ExpectedEnabled, cds.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, cds), failMsg)
	}
}
