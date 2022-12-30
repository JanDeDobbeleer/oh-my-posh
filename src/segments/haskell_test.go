package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/mock"
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"

	"github.com/stretchr/testify/assert"
)

func TestHaskell(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		GhcVersion      string
		StackGhcVersion string
		StackGhcMode    string
		InStackPackage  bool
		StackGhc        bool
	}{
		{
			Case:            "GHC 8.10.7",
			ExpectedString:  "8.10.7",
			GhcVersion:      "8.10.7",
			StackGhcVersion: "9.0.2",
			StackGhcMode:    "never",
		},
		{
			Case:            "Stack GHC Mode - Always",
			ExpectedString:  "9.0.2",
			GhcVersion:      "8.10.7",
			StackGhcVersion: "9.0.2",
			StackGhcMode:    "always",
			StackGhc:        true,
		},
		{
			Case:            "Stack GHC Mode - Package",
			ExpectedString:  "9.0.2",
			GhcVersion:      "8.10.7",
			StackGhcVersion: "9.0.2",
			StackGhcMode:    "package",
			InStackPackage:  true,
			StackGhc:        true,
		},
		{
			Case:            "Stack GHC Mode - Package no stack.yaml",
			ExpectedString:  "8.10.7",
			GhcVersion:      "8.10.7",
			StackGhcVersion: "9.0.2",
			StackGhcMode:    "package",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		if tc.StackGhcMode == "always" || (tc.StackGhcMode == "package" && tc.InStackPackage) {
			env.On("HasCommand", "stack").Return(true)
			env.On("RunCommand", "stack", []string{"ghc", "--", "--numeric-version"}).Return(tc.StackGhcVersion, nil)
		} else {
			env.On("HasCommand", "ghc").Return(true)
			env.On("RunCommand", "ghc", []string{"--numeric-version"}).Return(tc.GhcVersion, nil)
		}
		fileInfo := &platform.FileInfo{
			Path:         "../stack.yaml",
			ParentFolder: "./",
			IsDir:        false,
		}
		if tc.InStackPackage {
			var err error
			env.On("HasParentFilePath", "stack.yaml").Return(fileInfo, err)
		} else {
			env.On("HasParentFilePath", "stack.yaml").Return(fileInfo, errors.New("no match"))
		}
		env.On("HasFiles", "*.hs").Return(true)
		env.On("Pwd").Return("/usr/home/project")
		env.On("Home").Return("/usr/home")
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: make(map[string]string),
		})

		props := properties.Map{
			properties.FetchVersion: true,
		}
		props[StackGhcMode] = tc.StackGhcMode

		h := &Haskell{}
		h.Init(props, env)

		failMsg := fmt.Sprintf("Failed in case: %s", tc.Case)
		assert.True(t, h.Enabled(), failMsg)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, h.Template(), h), failMsg)
		assert.Equal(t, tc.StackGhc, h.StackGhc, failMsg)
	}
}
