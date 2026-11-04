package segments

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_mock "github.com/stretchr/testify/mock"
)

func TestUnreal(t *testing.T) {
	const (
		projectDir = "/home/dev/game"
		engineDir  = "/engines/UE_5.4"
		sourceDir  = "/src/UnrealEngine"
		buildFile  = "Engine/Build/Build.version"
		buildJSON  = `{"MajorVersion": 5, "MinorVersion": 4, "PatchVersion": 2, "Changelist": 1}`
	)

	cases := []struct {
		Files           map[string]string
		Registry        map[string]string
		RegistryErr     error
		Case            string
		Pwd             string
		GOOS            string
		ProjectFile     string
		Expected        string
		ExpectedEnabled bool
	}{
		{
			Case:            "launcher install resolved through the registry",
			Pwd:             projectDir,
			GOOS:            runtime.WINDOWS,
			ProjectFile:     `{"EngineAssociation": "5.4"}`,
			Registry:        map[string]string{`HKLM\SOFTWARE\EpicGames\Unreal Engine\5.4\InstalledDirectory`: engineDir},
			Files:           map[string]string{engineDir + "/" + buildFile: buildJSON},
			Expected:        "5.4.2",
			ExpectedEnabled: true,
		},
		{
			Case:            "source build resolved through the registry by GUID",
			Pwd:             projectDir,
			GOOS:            runtime.WINDOWS,
			ProjectFile:     `{"EngineAssociation": "{ABC-123}"}`,
			Registry:        map[string]string{`HKCU\Software\Epic Games\Unreal Engine\Builds\{ABC-123}`: engineDir},
			Files:           map[string]string{engineDir + "/" + buildFile: buildJSON},
			Expected:        "5.4.2",
			ExpectedEnabled: true,
		},
		{
			Case:            "found from a project subfolder",
			Pwd:             projectDir + "/Source/Module",
			GOOS:            runtime.WINDOWS,
			ProjectFile:     `{"EngineAssociation": "5.4"}`,
			Registry:        map[string]string{`HKLM\SOFTWARE\EpicGames\Unreal Engine\5.4\InstalledDirectory`: engineDir},
			Files:           map[string]string{engineDir + "/" + buildFile: buildJSON},
			Expected:        "5.4.2",
			ExpectedEnabled: true,
		},
		{
			Case:            "registry lookup fails, falls back to the association",
			Pwd:             projectDir,
			GOOS:            runtime.WINDOWS,
			ProjectFile:     `{"EngineAssociation": "5.4"}`,
			RegistryErr:     errors.New("access denied"),
			Expected:        "5.4",
			ExpectedEnabled: true,
		},
		{
			Case:            "Build.version missing, falls back to the association",
			Pwd:             projectDir,
			GOOS:            runtime.WINDOWS,
			ProjectFile:     `{"EngineAssociation": "5.4"}`,
			Registry:        map[string]string{`HKLM\SOFTWARE\EpicGames\Unreal Engine\5.4\InstalledDirectory`: engineDir},
			Expected:        "5.4",
			ExpectedEnabled: true,
		},
		{
			Case:            "Build.version malformed, falls back to the association",
			Pwd:             projectDir,
			GOOS:            runtime.WINDOWS,
			ProjectFile:     `{"EngineAssociation": "5.4"}`,
			Registry:        map[string]string{`HKLM\SOFTWARE\EpicGames\Unreal Engine\5.4\InstalledDirectory`: engineDir},
			Files:           map[string]string{engineDir + "/" + buildFile: "{"},
			Expected:        "5.4",
			ExpectedEnabled: true,
		},
		{
			Case:            "custom identifier is shown as is",
			Pwd:             projectDir,
			GOOS:            runtime.LINUX,
			ProjectFile:     `{"EngineAssociation": "my-fork"}`,
			Expected:        "my-fork",
			ExpectedEnabled: true,
		},
		{
			Case:            "non-Windows does not query the registry",
			Pwd:             projectDir,
			GOOS:            runtime.DARWIN,
			ProjectFile:     `{"EngineAssociation": "5.4"}`,
			Expected:        "5.4",
			ExpectedEnabled: true,
		},
		{
			Case:        "project without an association shows nothing",
			Pwd:         projectDir,
			GOOS:        runtime.WINDOWS,
			ProjectFile: `{}`,
		},
		{
			Case:        "malformed project file shows nothing",
			Pwd:         projectDir,
			GOOS:        runtime.WINDOWS,
			ProjectFile: `{`,
		},
		{
			Case:            "engine source checkout",
			Pwd:             sourceDir + "/Engine/Source",
			GOOS:            runtime.LINUX,
			Files:           map[string]string{sourceDir + "/" + buildFile: buildJSON},
			Expected:        "5.4.2",
			ExpectedEnabled: true,
		},
		{
			Case: "outside Unreal projects",
			Pwd:  "/home/dev/other",
			GOOS: runtime.LINUX,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			pwd := filepath.FromSlash(tc.Pwd)
			project := filepath.FromSlash(projectDir)

			env := new(mock.Environment)
			env.On("Pwd").Return(pwd)
			env.On("GOOS").Return(tc.GOOS)

			for dir := pwd; ; dir = filepath.Dir(dir) {
				isProject := tc.ProjectFile != "" && dir == project
				env.On("HasFilesInDir", dir, "*.uproject").Return(isProject)
				if dir == filepath.Dir(dir) {
					break
				}
			}

			if tc.ProjectFile != "" {
				env.On("LsDir", project).Return([]fs.DirEntry{&MockDirEntry{name: "Game.uproject"}})
				env.On("FileContent", filepath.Join(project, "Game.uproject")).Return(tc.ProjectFile)
			}

			for path, content := range tc.Files {
				env.On("FileContent", filepath.FromSlash(path)).Return(content)
			}
			env.On("FileContent", testify_mock.Anything).Return("")

			for key, value := range tc.Registry {
				env.On("WindowsRegistryKeyValue", key).Return(&runtime.WindowsRegistryValue{ValueType: runtime.STRING, String: filepath.FromSlash(value)}, nil)
			}
			env.On("WindowsRegistryKeyValue", testify_mock.Anything).Return(&runtime.WindowsRegistryValue{}, tc.RegistryErr)

			unreal := &Unreal{}
			unreal.Init(nil, env)

			assert.Equal(t, tc.ExpectedEnabled, unreal.Enabled())
			assert.Equal(t, tc.Expected, unreal.Version)
		})
	}
}
