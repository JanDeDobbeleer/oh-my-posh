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

func TestCFSegment(t *testing.T) {
	cases := []struct {
		Case            string
		Template        string
		ExpectedString  string
		ExpectedEnabled bool
		CfYamlFile      string
		Version         string
		DisplayMode     string
	}{
		{
			Case:            "1) cf 2.12.1 - file manifest.yml",
			ExpectedString:  "2.12.1",
			ExpectedEnabled: true,
			CfYamlFile:      "manifest.yml",
			Version:         `cf.exe version 2.12.1+645c3ce6a.2021-08-16`,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "2) cf 11.0.0-rc1 - file mta.yaml",
			Template:        "{{ .Major }}",
			ExpectedString:  "11",
			ExpectedEnabled: true,
			CfYamlFile:      "mta.yaml",
			Version:         `cf version 11.0.0-rc1`,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "3) cf 11.0.0-rc1 - no file",
			Template:        "{{ .Major }}",
			ExpectedString:  "",
			ExpectedEnabled: false,
			Version:         `cf version 11.0.0-rc1`,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "4) cf 11.1.0-rc1 - mode always",
			Template:        "{{ .Major }}.{{ .Minor }}",
			ExpectedString:  "11.1",
			ExpectedEnabled: true,
			Version:         `cf.exe version 11.1.0-rc1`,
			DisplayMode:     DisplayModeAlways,
		},
	}

	for _, tc := range cases {
		var env = new(mock.MockedEnvironment)
		env.On("HasCommand", "cf").Return(true)
		env.On("RunCommand", "cf", []string{"version"}).Return(tc.Version, nil)
		env.On("Pwd").Return("/usr/home/dev/my-app")
		env.On("Home").Return("/usr/home")

		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})

		cf := &Cf{}

		props := properties.Map{
			DisplayMode: tc.DisplayMode,
		}

		if tc.Template == "" {
			tc.Template = cf.Template()
		}

		cf.Init(props, env)

		for _, f := range cf.language.extensions {
			match, err := filepath.Match(f, tc.CfYamlFile)

			if err != nil {
				t.Fail()
			}

			if match {
				env.On("HasFiles", f).Return(true)
			} else {
				env.On("HasFiles", f).Return(false)
			}
		}

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.Equal(t, tc.ExpectedEnabled, cf.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, cf), failMsg)
	}
}
