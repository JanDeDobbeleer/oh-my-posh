package segments

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestPulumi(t *testing.T) {
	cases := []struct {
		Case       string
		YAMLConfig string
		JSONConfig string

		HasCommand bool

		FetchStack bool
		Stack      string
		StackError error

		HasWorkspaceFolder bool
		WorkSpaceFile      string

		FetchAbout bool
		About      string
		AboutError error
		AboutCache string

		ExpectedString  string
		ExpectedEnabled bool
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
		{
			Case:               "pulumi URL - cache",
			ExpectedString:     "\U000f0d46 1337 :: posh-user@s3://test-pulumi-state-test",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			FetchAbout:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
			AboutCache:         `{ "url": "s3://test-pulumi-state-test", "user":"posh-user" }`,
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
			Case:               "pulumi URL - cache error",
			ExpectedString:     "\U000f0d46 1337 :: posh-user@s3://test-pulumi-state-test-output",
			ExpectedEnabled:    true,
			HasCommand:         true,
			HasWorkspaceFolder: true,
			FetchStack:         true,
			FetchAbout:         true,
			JSONConfig:         `{ "name": "oh-my-posh" }`,
			WorkSpaceFile:      `{ "stack": "1337" }`,
			AboutCache:         `{`,
			About:              `{ "backend": { "url": "s3://test-pulumi-state-test-output", "user":"posh-user" } }`,
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
		env := new(mock.MockedEnvironment)

		env.On("HasCommand", "pulumi").Return(tc.HasCommand)
		env.On("RunCommand", "pulumi", []string{"stack", "ls", "--json"}).Return(tc.Stack, tc.StackError)
		env.On("RunCommand", "pulumi", []string{"about", "--json"}).Return(tc.About, tc.AboutError)

		env.On("Pwd").Return("/home/foobar/Work/oh-my-posh/pulumi/projects/awesome-project")
		env.On("Home").Return(filepath.Clean("/home/foobar"))
		env.On("Error", mock2.Anything)
		env.On("Debug", mock2.Anything)
		env.On("DebugF", mock2.Anything, mock2.Anything)

		env.On("HasFiles", pulumiYAML).Return(len(tc.YAMLConfig) > 0)
		env.On("FileContent", pulumiYAML).Return(tc.YAMLConfig, nil)

		env.On("HasFiles", pulumiJSON).Return(len(tc.JSONConfig) > 0)
		env.On("FileContent", pulumiJSON).Return(tc.JSONConfig, nil)

		env.On("PathSeparator").Return("/")

		env.On("HasFolder", filepath.Clean("/home/foobar/.pulumi/workspaces")).Return(tc.HasWorkspaceFolder)
		workspaceFile := "oh-my-posh-c62b7b6786c5c5a85896576e46a25d7c9f888e92-workspace.json"
		env.On("HasFilesInDir", filepath.Clean("/home/foobar/.pulumi/workspaces"), workspaceFile).Return(len(tc.WorkSpaceFile) > 0)
		env.On("FileContent", filepath.Clean("/home/foobar/.pulumi/workspaces/"+workspaceFile)).Return(tc.WorkSpaceFile, nil)

		cache := &mock.MockedCache{}
		cache.On("Get", "pulumi-oh-my-posh-1337-c62b7b6786c5c5a85896576e46a25d7c9f888e92-about").Return(tc.AboutCache, len(tc.AboutCache) > 0)
		cache.On("Set", mock2.Anything, mock2.Anything, mock2.Anything)

		env.On("Cache").Return(cache)

		pulumi := &Pulumi{
			env: env,
			props: properties.Map{
				FetchStack: tc.FetchStack,
				FetchAbout: tc.FetchAbout,
			},
		}

		assert.Equal(t, tc.ExpectedEnabled, pulumi.Enabled(), tc.Case)

		if !tc.ExpectedEnabled {
			continue
		}

		var got = renderTemplate(env, pulumi.Template(), pulumi)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
