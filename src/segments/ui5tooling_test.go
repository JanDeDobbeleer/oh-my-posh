package segments

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

const (
	WorkingDirRoot = "/home/user/dev/my-app"
)

type testCase struct {
	Case            string
	Template        string
	ExpectedString  string
	ExpectedEnabled bool
	UI5YamlFilename string
	WorkingDir      string
	Version         string
	DisplayMode     string
}

func TestUI5Tooling(t *testing.T) {
	cases := []testCase{
		{
			Case:            "1) ui5tooling 2.12.1 - file ui5.yaml present in cwd; DisplayMode = files",
			ExpectedString:  "2.12.1",
			ExpectedEnabled: true,
			UI5YamlFilename: "ui5.yaml",
			Version:         `2.12.1 (from C:\somewhere\cli\bin\ui5.js)`,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "2) ui5tooling 2.12.2 - file ui5.yaml present in cwd; default display mode (context)",
			ExpectedString:  "2.12.2",
			ExpectedEnabled: true,
			UI5YamlFilename: "ui5.yaml",
			Version:         `2.12.2 (from C:\somewhere\cli\bin\ui5.js)`,
		},
		{
			Case:            "3) ui5tooling 2.12.3 - file ui5.yaml present; cwd is sub dir, default display mode (context)",
			ExpectedString:  "2.12.3",
			WorkingDir:      WorkingDirRoot + "/subdir",
			ExpectedEnabled: true,
			UI5YamlFilename: "ui5.yaml",
			Version:         `2.12.3 (from C:\somewhere\cli\bin\ui5.js)`,
		},
		{
			Case:            "4) no ui5tooling segment - file ui5.yaml present, cwd is sub dir; display mode = files",
			ExpectedString:  "",
			WorkingDir:      WorkingDirRoot + "/subdir",
			ExpectedEnabled: false,
			UI5YamlFilename: "ui5.yaml",
			DisplayMode:     DisplayModeFiles,
			Version:         `2.12.1 (from C:\somewhere\cli\bin\ui5.js)`,
		},
		{
			Case:            "5) ui5tooling 2.12.4 - file ui5-dist.yml present in cwd",
			ExpectedString:  "2.12.4",
			ExpectedEnabled: true,
			UI5YamlFilename: "ui5-dist.yml",
			Version:         `2.12.4 (from C:\somewhere\cli\bin\ui5.js)`,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "6) no ui5tooling segment - file ui5.yaml not present, display mode = files",
			ExpectedString:  "",
			ExpectedEnabled: false,
			Version:         `2.12.1 (from C:\somewhere\cli\bin\ui5.js)`,
			DisplayMode:     DisplayModeFiles,
		},
		{
			Case:            "7) no ui5tooling segment - file ui5.yaml not present, default display mode (context)",
			ExpectedString:  "",
			ExpectedEnabled: false,
			Version:         `2.12.1 (from C:\somewhere\cli\bin\ui5.js)`,
		},
		{
			Case:            "8) ui5tooling 11.0.0-rc1, no ui5.yaml file but display mode = always",
			Template:        "{{ .Major }}",
			ExpectedString:  "11",
			ExpectedEnabled: true,
			Version:         `11.0.0-rc1 (from C:\somewhere\cli\bin\ui5.js)`,
			DisplayMode:     DisplayModeAlways,
		},
	}

	for _, tc := range cases {
		env := prepareMockedEnvironment(&tc)
		ui5tooling := &UI5Tooling{}

		if tc.WorkingDir == "" {
			tc.WorkingDir = WorkingDirRoot
		}

		if tc.DisplayMode == "" {
			tc.DisplayMode = DisplayModeContext
		}

		if tc.Template == "" {
			tc.Template = ui5tooling.Template()
		}

		props := properties.Map{
			DisplayMode: tc.DisplayMode,
		}

		ui5tooling.Init(props, env)
		err := mockFilePresence(&tc, ui5tooling, env)

		if err != nil {
			t.Fail()
		}

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.Equal(t, tc.ExpectedEnabled, ui5tooling.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, ui5tooling), failMsg)
	}
}

func prepareMockedEnvironment(tc *testCase) *mock.MockedEnvironment {
	var env = new(mock.MockedEnvironment)
	env.On("HasCommand", "ui5").Return(true)
	env.On("RunCommand", "ui5", []string{"--version"}).Return(tc.Version, nil)
	env.On("Home").Return("/home/user")
	env.On("Pwd").Return(WorkingDirRoot)

	env.On("TemplateCache").Return(&platform.TemplateCache{
		Env: make(map[string]string),
	})

	return env
}

func mockFilePresence(tc *testCase, ui5tooling *UI5Tooling, env *mock.MockedEnvironment) error {
	for _, f := range ui5tooling.language.extensions {
		match, err := filepath.Match(f, tc.UI5YamlFilename)

		if err != nil {
			return err
		}

		if match {
			if tc.DisplayMode == DisplayModeFiles && tc.WorkingDir == WorkingDirRoot {
				env.On("HasFiles", f).Return(true)
				env.On("HasFileInParentDirs", f, uint(4)).Return(false)
				// mode context, working dir != working dir root
			} else if tc.DisplayMode == DisplayModeContext {
				env.On("HasFileInParentDirs", f, uint(4)).Return(false)
				env.On("HasFiles", f).Return(true)
			} else {
				env.On("HasFileInParentDirs", f, uint(4)).Return(false)
				env.On("HasFiles", f).Return(false)
			}
		} else {
			env.On("HasFileInParentDirs", f, uint(4)).Return(false)
			env.On("HasFiles", f).Return(false)
		}
	}

	return nil
}
