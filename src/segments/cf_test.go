package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCFSegment(t *testing.T) {
	cases := []struct {
		Case           string
		Template       string
		ExpectedString string
		CfYamlFile     string
		Version        string
		DisplayMode    string
	}{
		{
			Case:           "1) cf 2.12.1 - file manifest.yml",
			ExpectedString: "2.12.1",
			CfYamlFile:     "manifest.yml",
			Version:        `cf.exe version 2.12.1+645c3ce6a.2021-08-16`,
			DisplayMode:    DisplayModeFiles,
		},
		{
			Case:           "2) cf 11.0.0-rc1 - file mta.yaml",
			Template:       "{{ .Major }}",
			ExpectedString: "11",
			CfYamlFile:     "mta.yaml",
			Version:        `cf version 11.0.0-rc1`,
			DisplayMode:    DisplayModeFiles,
		},
		{
			Case:           "4) cf 11.1.0-rc1 - mode always",
			Template:       "{{ .Major }}.{{ .Minor }}",
			ExpectedString: "11.1",
			Version:        `cf.exe version 11.1.0-rc1`,
			DisplayMode:    DisplayModeAlways,
		},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "cf",
			versionParam:  "version",
			versionOutput: tc.Version,
			extension:     "manifest.yml",
		}
		env, props := getMockedLanguageEnv(params)

		props[DisplayMode] = tc.DisplayMode

		cf := &Cf{}
		cf.Init(props, env)

		if tc.Template == "" {
			tc.Template = cf.Template()
		}

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, cf.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, cf), failMsg)
	}
}
