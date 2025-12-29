package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestWinGet(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedCount   int
		WinGetOutput    string
		CommandError    error
		GOOS            string
		HasCommand      bool
	}{
		{
			Case:            "No updates available",
			ExpectedEnabled: false,
			ExpectedCount:   0,
			GOOS:            runtime.WINDOWS,
			HasCommand:      true,
			WinGetOutput: `Name               Id                          Version   Available Source
-----------------------------------------------------------------------------------
No applicable updates found.`,
		},
		{
			Case:            "Multiple updates available",
			ExpectedEnabled: true,
			ExpectedCount:   3,
			GOOS:            runtime.WINDOWS,
			HasCommand:      true,
			WinGetOutput: `Name               Id                          Version   Available Source
-----------------------------------------------------------------------------------
Python 3.11        Python.Python.3.11          3.11.0    3.11.5    winget
Node.js            OpenJS.NodeJS               18.0.0    18.12.1   winget
Git                Git.Git                     2.39.0    2.40.0    winget
3 upgrades available.`,
		},
		{
			Case:            "Single update available",
			ExpectedEnabled: true,
			ExpectedCount:   1,
			GOOS:            runtime.WINDOWS,
			HasCommand:      true,
			WinGetOutput: `Name               Id                          Version   Available Source
-----------------------------------------------------------------------------------
Python 3.11        Python.Python.3.11          3.11.0    3.11.5    winget`,
		},
		{
			Case:            "Non-Windows OS",
			ExpectedEnabled: false,
			GOOS:            runtime.LINUX,
			HasCommand:      true,
		},
		{
			Case:            "WinGet command not found",
			ExpectedEnabled: false,
			GOOS:            runtime.WINDOWS,
			HasCommand:      false,
		},
		{
			Case:            "Command execution error",
			ExpectedEnabled: false,
			GOOS:            runtime.WINDOWS,
			HasCommand:      true,
			CommandError:    &runtime.CommandError{Err: "command failed"},
		},
		{
			Case:            "Updates with Unicode separator",
			ExpectedEnabled: true,
			ExpectedCount:   2,
			GOOS:            runtime.WINDOWS,
			HasCommand:      true,
			WinGetOutput: `Name               Id                          Version   Available Source
─────────────────────────────────────────────────────────────────────────────────
Docker Desktop     Docker.DockerDesktop        4.16.0    4.17.0    winget
Visual Studio Code Microsoft.VisualStudioCode  1.75.0    1.76.0    winget`,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return(tc.GOOS)
		env.On("HasCommand", "winget").Return(tc.HasCommand)

		if tc.CommandError != nil {
			env.On("RunCommand", "winget", []string{"upgrade"}).Return("", tc.CommandError)
		} else if tc.HasCommand && tc.GOOS == runtime.WINDOWS {
			env.On("RunCommand", "winget", []string{"upgrade"}).Return(tc.WinGetOutput, nil)
		}

		cache.DeleteAll(cache.Device)

		w := &WinGet{}
		w.Init(options.Map{}, env)

		enabled := w.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if enabled {
			assert.Equal(t, tc.ExpectedCount, w.UpdateCount, tc.Case)
			assert.Equal(t, tc.ExpectedCount, len(w.Updates), tc.Case)
		}
	}
}

func TestWinGetParseOutput(t *testing.T) {
	cases := []struct {
		Case          string
		Output        string
		ExpectedCount int
		ExpectedFirst WinGetPackage
	}{
		{
			Case: "Standard output",
			Output: `Name               Id                          Version   Available Source
-----------------------------------------------------------------------------------
Python 3.11        Python.Python.3.11          3.11.0    3.11.5    winget
Node.js            OpenJS.NodeJS               18.0.0    18.12.1   winget`,
			ExpectedCount: 2,
			ExpectedFirst: WinGetPackage{
				Name:      "Python 3.11",
				ID:        "Python.Python.3.11",
				Current:   "3.11.0",
				Available: "3.11.5",
			},
		},
		{
			Case: "Empty output",
			Output: `Name               Id                          Version   Available Source
-----------------------------------------------------------------------------------`,
			ExpectedCount: 0,
		},
		{
			Case: "Output with footer",
			Output: `Name               Id                          Version   Available Source
-----------------------------------------------------------------------------------
Python 3.11        Python.Python.3.11          3.11.0    3.11.5    winget
2 upgrades available.`,
			ExpectedCount: 1,
			ExpectedFirst: WinGetPackage{
				Name:      "Python 3.11",
				ID:        "Python.Python.3.11",
				Current:   "3.11.0",
				Available: "3.11.5",
			},
		},
	}

	for _, tc := range cases {
		w := &WinGet{}
		packages := w.parseWinGetOutput(tc.Output)

		assert.Equal(t, tc.ExpectedCount, len(packages), tc.Case)
		if tc.ExpectedCount > 0 {
			assert.Equal(t, tc.ExpectedFirst.Name, packages[0].Name, tc.Case)
		}
	}
}
