package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/stretchr/testify/assert"
)

func TestPulumi(t *testing.T) {
	cases := []struct {
		StackError         error
		AboutError         error
		About              string
		YAMLConfig         string
		JSONConfig         string
		Case               string
		ExpectedString     string
		Stack              string
		AboutCache         string
		WorkSpaceFile      string
		HasCommand         bool
		FetchAbout         bool
		HasWorkspaceFolder bool
		FetchStack         bool
		ExpectedEnabled    bool
	}{
		{
			Case:            "no pulumi command",
			ExpectedEnabled: false,
			HasCommand:      false,
		},
		{
			Case:            "pulumi command is present, but no pulumi file",
			ExpectedEnabled: false,
			HasCommand:      true,
		},
		{
			Case:            "pulumi file YAML is present",
			ExpectedString:  "\U000f0d46",
			ExpectedEnabled: true,
			HasCommand:      true,
			YAMLConfig: `
name: oh-my-posh
runtime: golang
description: A Console App
`,
		},
		{
			Case:            "pulumi file JSON is present",
			ExpectedString:  "\U000f0d46",
			ExpectedEnabled: true,
			HasCommand:      true,
			JSONConfig:      `{ "name": "oh-my-posh" }`,
		},
		{
			Case:               "no stack present",
			ExpectedString:     "\U000f0d46 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
		},
		{
			Case:               "pulumi stack",
			ExpectedString:     "\U000f0d46 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
		},
		{
			Case:               "pulumi URL",
			ExpectedString:     "\U000f0d46 1337 :: posh-user@s3://test-pulumi-state-test",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			FetchAbout:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
			About:              `{ "backend": { "url": "s3://test-pulumi-state-test", "user":"posh-user" } }`,
		},
		// Error flows
		{
			Case:            "pulumi file JSON error",
			ExpectedString:  "\U000f0d46",
			ExpectedEnabled: true,
			FetchStack:      true,
			HasCommand:      true,
			JSONConfig:      `{`,
		},
		{
			Case:               "pulumi workspace file JSON error",
			ExpectedString:     "\U000f0d46",
			ExpectedEnabled:    true,
			FetchStack:         true,
			HasCommand:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{`,
			HasWorkspaceFolder: true,
		},
		{
			Case:            "pulumi URL, no fetch_stack set",
			ExpectedString:  "\U000f0d46",
			ExpectedEnabled: true,
			HasCommand:      true,
			FetchAbout:      true,
			JSONConfig:      `{ "name": "oh-my-posh" }`,
		},
		{
			Case:               "pulumi URL - about error",
			ExpectedString:     "\U000f0d46 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			FetchAbout:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
			AboutError:         errors.New("error"),
		},
		{
			Case:               "pulumi URL - about decode error",
			ExpectedString:     "\U000f0d46 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			FetchAbout:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
			About:              `{`,
		},
		{
			Case:               "pulumi URL - about backend is nil",
			ExpectedString:     "\U000f0d46 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			FetchAbout:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
			About:              `{}`,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		env.On("HasCommand", "pulumi").Return(tc.HasCommand)
		env.On("RunCommand", "pulumi", []string{"stack", "ls", "--json"}).Return(tc.Stack, tc.StackError)
		env.On("RunCommand", "pulumi", []string{"about", "--json"}).Return(tc.About, tc.AboutError)

		pwd := "/home/foobar/Work/oh-my-posh/pulumi/projects/awesome-project"
		env.On("Pwd").Return(pwd)
		env.On("Home").Return(filepath.Clean("/home/foobar"))

		env.On("HasFiles", pulumiYAML).Return(len(tc.YAMLConfig) > 0)
		env.On("FileContent", pulumiYAML).Return(tc.YAMLConfig, nil)

		env.On("HasFiles", pulumiJSON).Return(len(tc.JSONConfig) > 0)
		env.On("FileContent", pulumiJSON).Return(tc.JSONConfig, nil)

		env.On("HasFolder", filepath.Clean("/home/foobar/.pulumi/workspaces")).Return(tc.HasWorkspaceFolder)

		pulumi := &Pulumi{}

		var fileName string
		if len(tc.JSONConfig) > 0 {
			fileName = pulumiJSON
		} else {
			fileName = pulumiYAML
		}

		sha1 := pulumi.sha1HexString(pwd + path.Separator() + fileName)
		workspaceFile := fmt.Sprintf("oh-my-posh-%s-workspace.json", sha1)

		env.On("HasFilesInDir", filepath.Clean("/home/foobar/.pulumi/workspaces"), workspaceFile).Return(len(tc.WorkSpaceFile) > 0)

		env.On("FileContent", filepath.Clean("/home/foobar/.pulumi/workspaces/"+workspaceFile)).Return(tc.WorkSpaceFile, nil)

		props := properties.Map{
			FetchStack: tc.FetchStack,
			FetchAbout: tc.FetchAbout,
		}

		pulumi.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, pulumi.Enabled(), tc.Case)

		if !tc.ExpectedEnabled {
			continue
		}

		var got = renderTemplate(env, pulumi.Template(), pulumi)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
