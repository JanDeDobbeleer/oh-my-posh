package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCdsSegment(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Template       string
		Version        string
		PackageJSON    string
		DisplayMode    string
	}{
		{
			Case:           "1) cds 5.5.0 - file .cdsrc.json present",
			ExpectedString: "5.5.0",
			Version:        "@sap/cds: 5.5.0\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			DisplayMode:    DisplayModeFiles,
		},
		{
			Case:           "2) cds 5.5.1 - file some.cds",
			ExpectedString: "5.5.1",
			Version:        "@sap/cds: 5.5.1\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			DisplayMode:    DisplayModeFiles,
		},
		{
			Case:           "4) cds 5.5.3 - package.json dependency",
			ExpectedString: "5.5.3",
			Version:        "@sap/cds: 5.5.3\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:    "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/cds\": \"^5\" } }",
			DisplayMode:    DisplayModeContext,
		},
		{
			Case:           "4) cds 5.5.4 - package.json dependency, major + minor",
			ExpectedString: "5.5",
			Template:       "{{ .Major }}.{{ .Minor }}",
			Version:        "@sap/cds: 5.5.4\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:    "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/cds\": \"^5\" } }",
			DisplayMode:    DisplayModeContext,
		},
		{
			Case:           "6) cds 5.5.9 - display always",
			ExpectedString: "5.5.9",
			Version:        "@sap/cds: 5.5.9\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			PackageJSON:    "{ \"name\": \"my-app\",\"dependencies\": { \"@sap/cds\": \"^5\" } }",
			DisplayMode:    DisplayModeAlways,
		},
		{
			Case:           "8) cds 5.5.0 - file .cdsrc-private.json present",
			ExpectedString: "5.5.0",
			Version:        "@sap/cds: 5.5.0\n@sap/cds-compiler: 2.7.0\n@sap/cds-dk: 4.5.3",
			DisplayMode:    DisplayModeFiles,
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "cds",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     ".cdsrc.json",
		}
		env, props := getMockedLanguageEnv(params)

		if len(tc.DisplayMode) == 0 {
			tc.DisplayMode = DisplayModeContext
		}
		props[DisplayMode] = tc.DisplayMode

		if len(tc.PackageJSON) != 0 {
			env.On("HasFiles", "package.json").Return(true)
			env.On("FileContent", "package.json").Return(tc.PackageJSON)
		} else {
			env.On("HasFiles", "package.json").Return(false)
		}

		cds := &Cds{}
		cds.Init(props, env)

		if tc.Template == "" {
			tc.Template = cds.Template()
		}

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, cds.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, cds), failMsg)
	}
}
