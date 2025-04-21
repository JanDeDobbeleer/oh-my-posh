package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/constants"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/stretchr/testify/assert"
)

func TestDotnetSegment(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Version  string
		ExitCode int
	}{
		{Case: "Unsupported version", Expected: "\uf071", ExitCode: constants.DotnetExitCode, Version: "3.1.402"},
		{Case: "Regular version", Expected: "3.1.402", Version: "3.1.402"},
	}

	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "dotnet",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.cs",
		}
		env, props := getMockedLanguageEnv(params)

		if tc.ExitCode != 0 {
			env.Unset("RunCommand")
			err := &runtime.CommandError{ExitCode: tc.ExitCode}
			env.On("RunCommand", "dotnet", []string{"--version"}).Return("", err)
		}

		dotnet := &Dotnet{}
		dotnet.Init(props, env)

		assert.True(t, dotnet.Enabled())
		assert.Equal(t, tc.Expected, renderTemplate(env, dotnet.Template(), dotnet), tc.Case)
	}
}

func TestDotnetSDKVersion(t *testing.T) {
	cases := []struct {
		Case           string
		GlobalJSON     string
		ExpectedSDK    string
		GlobalJSONPath string
		FetchSDK       bool
		HasGlobalJSON  bool
	}{
		{
			Case:        "Do not fetch SDK version",
			FetchSDK:    false,
			ExpectedSDK: "",
		},
		{
			Case:        "No global.json found",
			FetchSDK:    true,
			ExpectedSDK: "",
		},
		{
			Case:           "Valid global.json",
			FetchSDK:       true,
			GlobalJSON:     `{"sdk": {"version": "6.0.100"}}`,
			ExpectedSDK:    "6.0.100",
			HasGlobalJSON:  true,
			GlobalJSONPath: "/test/global.json",
		},
		{
			Case:           "Invalid global.json",
			FetchSDK:       true,
			GlobalJSON:     `invalid json`,
			ExpectedSDK:    "",
			HasGlobalJSON:  true,
			GlobalJSONPath: "/test/global.json",
		},
	}

	params := &mockedLanguageParams{
		cmd:           "dotnet",
		versionParam:  "--version",
		versionOutput: "6.0.100",
		extension:     "*.cs",
	}

	for _, tc := range cases {
		props := properties.Map{
			FetchSDKVersion:         tc.FetchSDK,
			properties.FetchVersion: false,
		}

		env, _ := getMockedLanguageEnv(params)

		if tc.HasGlobalJSON {
			file := &runtime.FileInfo{
				Path: tc.GlobalJSONPath,
			}
			env.On("HasParentFilePath", "global.json", false).Return(file, nil)
			env.On("FileContent", tc.GlobalJSONPath).Return(tc.GlobalJSON)
		} else {
			env.On("HasParentFilePath", "global.json", false).Return(&runtime.FileInfo{}, errors.New("file not found"))
		}

		dotnet := &Dotnet{}
		dotnet.Init(props, env)

		assert.True(t, dotnet.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedSDK, dotnet.SDKVersion, tc.Case)
	}
}
