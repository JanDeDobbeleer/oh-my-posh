package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestPulumi(t *testing.T) {
	cases := []struct {
		StackError                error
		AboutError                error
		About                     string
		YAMLConfig                string
		JSONConfig                string
		Case                      string
		ExpectedString            string
		Stack                     string
		AboutCache                string
		PulumiHome                string
		WorkSpaceFile             string
		DefaultWorkSpaceFile      string
		HasCommand                bool
		FetchAbout                bool
		HasWorkspaceFolder        bool
		HasDefaultWorkspaceFolder bool
		FetchStack                bool
		ExpectedEnabled           bool
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
			ExpectedString:  "\ue873",
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
			ExpectedString:  "\ue873",
			ExpectedEnabled: true,
			HasCommand:      true,
			JSONConfig:      `{ "name": "oh-my-posh" }`,
		},
		{
			Case:               "missing workspace file",
			ExpectedString:     "\ue873",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
		},
		{
			Case:               "pulumi stack in default home for unset or empty PULUMI_HOME",
			ExpectedString:     "\ue873 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
		},
		{
			Case:               "pulumi stack in custom absolute home",
			ExpectedString:     "\ue873 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			PulumiHome:         filepath.Clean("/custom/pulumi"),
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
		},
		{
			Case:               "pulumi stack in custom relative home",
			ExpectedString:     "\ue873 1337",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			PulumiHome:         filepath.Join("project", ".pulumi"),
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
		},
		{
			Case:                      "custom home does not fall back to default home",
			ExpectedString:            "\ue873",
			ExpectedEnabled:           true,
			HasCommand:                true,
			HasWorkspaceFolder:        true,
			HasDefaultWorkspaceFolder: true,
			FetchStack:                true,
			PulumiHome:                filepath.Clean("/custom/pulumi"),
			JSONConfig:                `{ "name": "oh-my-posh" }`,
			DefaultWorkSpaceFile:      `{ "stack": "stale" }`,
		},
		{
			Case:               "pulumi URL",
			ExpectedString:     "\ue873 1337 :: posh-user@s3://test-pulumi-state-test",
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
			ExpectedString:  "\ue873",
			ExpectedEnabled: true,
			FetchStack:      true,
			HasCommand:      true,
			JSONConfig:      `{`,
		},
		{
			Case:               "pulumi workspace file JSON error",
			ExpectedString:     "\ue873",
			ExpectedEnabled:    true,
			FetchStack:         true,
			HasCommand:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{`,
			HasWorkspaceFolder: true,
		},
		{
			Case:            "pulumi URL, no fetch_stack set",
			ExpectedString:  "\ue873",
			ExpectedEnabled: true,
			HasCommand:      true,
			FetchAbout:      true,
			JSONConfig:      `{ "name": "oh-my-posh" }`,
		},
		{
			Case:               "pulumi URL - about error",
			ExpectedString:     "\ue873 1337",
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
			ExpectedString:     "\ue873 1337",
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
			ExpectedString:     "\ue873 1337",
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
		home := filepath.Clean("/home/foobar")
		env.On("Home").Return(home)
		env.On("Getenv", "PULUMI_HOME").Return(tc.PulumiHome)

		env.On("HasFiles", pulumiYAML).Return(len(tc.YAMLConfig) > 0)
		env.On("FileContent", pulumiYAML).Return(tc.YAMLConfig, nil)

		env.On("HasFiles", pulumiJSON).Return(len(tc.JSONConfig) > 0)
		env.On("FileContent", pulumiJSON).Return(tc.JSONConfig, nil)

		pulumi := &Pulumi{}

		var fileName string
		if len(tc.JSONConfig) > 0 {
			fileName = pulumiJSON
		} else {
			fileName = pulumiYAML
		}

		sha1 := pulumi.sha1HexString(pwd + path.Separator() + fileName)
		workspaceFile := fmt.Sprintf("oh-my-posh-%s-workspace.json", sha1)
		defaultWorkspaceCacheDir := filepath.Join(home, ".pulumi", "workspaces")
		workspaceCacheDir := defaultWorkspaceCacheDir
		if tc.PulumiHome != "" {
			workspaceCacheDir = filepath.Join(tc.PulumiHome, "workspaces")
		}

		env.On("HasFolder", workspaceCacheDir).Return(tc.HasWorkspaceFolder)
		env.On("HasFilesInDir", workspaceCacheDir, workspaceFile).Return(len(tc.WorkSpaceFile) > 0)
		env.On("FileContent", filepath.Join(workspaceCacheDir, workspaceFile)).Return(tc.WorkSpaceFile)

		if workspaceCacheDir != defaultWorkspaceCacheDir {
			env.On("HasFolder", defaultWorkspaceCacheDir).Return(tc.HasDefaultWorkspaceFolder)
			env.On("HasFilesInDir", defaultWorkspaceCacheDir, workspaceFile).Return(len(tc.DefaultWorkSpaceFile) > 0)
			env.On("FileContent", filepath.Join(defaultWorkspaceCacheDir, workspaceFile)).Return(tc.DefaultWorkSpaceFile)
		}

		props := options.Map{
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
