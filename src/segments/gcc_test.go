package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
)

func TestGcc(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
		Files          []string
		Enabled        bool
	}{
		{
			Case:           "GCC 11.2.0 with C file",
			ExpectedString: "\ue7e5 11.2.0",
			Version:        "gcc (GCC) 11.2.0\nCopyright (C) 2021 Free Software Foundation, Inc.",
			Files:          []string{"*.c"},
			Enabled:        true,
		},
		{
			Case:           "GCC 9.3.0 with CPP file",
			ExpectedString: "\ue7e5 9.3.0",
			Version:        "gcc (GCC) 9.3.0\nCopyright (C) 2019 Free Software Foundation, Inc.",
			Files:          []string{"*.cpp"},
			Enabled:        true,
		},
		{
			Case:           "GCC 10.1.0 with CMakeLists.txt",
			ExpectedString: "\ue7e5 10.1.0",
			Version:        "gcc (GCC) 10.1.0\nCopyright (C) 2020 Free Software Foundation, Inc.",
			Files:          []string{"CMakeLists.txt"},
			Enabled:        true,
		},
		{
			Case:           "GCC 8.1.0 with header file",
			ExpectedString: "\ue7e5 8.1.0",
			Version:        "gcc (GCC) 8.1.0\nCopyright (C) 2018 Free Software Foundation, Inc.",
			Files:          []string{"*.h"},
			Enabled:        true,
		},
		{
			Case:           "No C/C++ files present",
			ExpectedString: "",
			Version:        "gcc (GCC) 11.2.0",
			Files:          []string{},
			Enabled:        false,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HasCommand", gccToolName).Return(tc.Enabled)

		// Mock version output
		env.On("RunCommandWithEnv", gccToolName, []string(nil), []string{versionFlagArg}).Return(tc.Version, nil)

		// Mock file checks
		env.On("HasFiles", "*.c").Return(contains(tc.Files, "*.c"))
		env.On("HasFiles", "*.cpp").Return(contains(tc.Files, "*.cpp"))
		env.On("HasFiles", "*.h").Return(contains(tc.Files, "*.h"))
		env.On("HasFiles", "CMakeLists.txt").Return(contains(tc.Files, "CMakeLists.txt"))

		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("Shell").Return("pwsh")

		if template.Cache == nil {
			template.Cache = &cache.Template{}
		}
		template.Init(env, nil, nil)

		props := options.Map{
			options.FetchVersion: true,
		}

		gcc := &Gcc{}
		gcc.Init(props, env)

		assert.Equal(t, tc.Enabled, gcc.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		if tc.Enabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, gcc.Template(), gcc), fmt.Sprintf("Failed in case: %s", tc.Case))
		}
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
