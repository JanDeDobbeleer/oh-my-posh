package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/alecthomas/assert"
)

func TestMvn(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		MvnVersion     string
		MvnwVersion    string
		HasMvnw        bool
	}{
		{
			Case:           "Maven version",
			ExpectedString: "1.0.0",
			MvnVersion:     "Apache Maven 1.0.0",
			HasMvnw:        false,
			MvnwVersion:    ""},
		{
			Case:           "Local Maven version from wrapper",
			ExpectedString: "1.1.0-beta-9",
			MvnVersion:     "Apache Maven 1.0.0",
			HasMvnw:        true,
			MvnwVersion:    "Apache Maven 1.1.0-beta-9"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "mvn",
			versionParam:  "--version",
			versionOutput: tc.MvnVersion,
			extension:     "pom.xml",
		}
		env, props := getMockedLanguageEnv(params)

		fileInfo := &runtime.FileInfo{
			Path:         "../mvnw",
			ParentFolder: "./",
			IsDir:        false,
		}
		var err error
		if !tc.HasMvnw {
			err = errors.New("no match")
		}
		env.On("HasParentFilePath", "mvnw", false).Return(fileInfo, err)
		env.On("HasCommand", fileInfo.Path).Return(tc.HasMvnw)
		env.On("RunCommand", fileInfo.Path, []string{"--version"}).Return(tc.MvnwVersion, nil)

		m := &Mvn{}
		m.Init(props, env)
		assert.True(t, m.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, m.Template(), m), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
