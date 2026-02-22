package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestUnoSegment(t *testing.T) {
	cases := []struct {
		Case            string
		GlobalJSON      string
		ExpectedVersion string
		ExpectedEnabled bool
		HasGlobalJSON   bool
	}{
		{
			Case:            "no global.json found",
			ExpectedEnabled: false,
		},
		{
			Case:            "invalid global.json",
			HasGlobalJSON:   true,
			GlobalJSON:      `invalid json`,
			ExpectedEnabled: false,
		},
		{
			Case:            "global.json without Uno.Sdk",
			HasGlobalJSON:   true,
			GlobalJSON:      `{"msbuild-sdks":{"Another.Sdk":"1.0.0"}}`,
			ExpectedEnabled: false,
		},
		{
			Case:            "valid Uno.Sdk in global.json",
			HasGlobalJSON:   true,
			GlobalJSON:      `{"msbuild-sdks":{"Uno.Sdk":"6.5.31"}}`,
			ExpectedVersion: "6.5.31",
			ExpectedEnabled: true,
		},
		{
			Case:            "valid dotnet sdk global.json without Uno.Sdk",
			HasGlobalJSON:   true,
			GlobalJSON:      `{"sdk":{ "version": "10.0.103" }}`,
			ExpectedEnabled: false,
		},
		{
			Case:          "valid Uno.Sdk in commented global.json",
			HasGlobalJSON: true,
			GlobalJSON: `{
  // Uno SDK version
  "msbuild-sdks": {
    "Uno.Sdk": "6.5.31"
  }
}`,
			ExpectedVersion: "6.5.31",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		if tc.HasGlobalJSON {
			file := &runtime.FileInfo{Path: "/test/global.json"}
			env.On("HasParentFilePath", "global.json", false).Return(file, nil)
			env.On("FileContent", "/test/global.json").Return(tc.GlobalJSON)
		} else {
			env.On("HasParentFilePath", "global.json", false).Return(&runtime.FileInfo{}, errors.New("file not found"))
		}

		uno := &Uno{}
		uno.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, uno.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedVersion, uno.Version, tc.Case)

		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedVersion, renderTemplate(env, "{{ .Version }}", uno), tc.Case)
		}
	}
}
