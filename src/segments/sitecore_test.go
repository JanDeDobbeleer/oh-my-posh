package segments

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"

	"github.com/stretchr/testify/assert"
)

func TestSitecoreSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		Template        string
		IsCloud         bool
	}{
		{
			Case:            "no config files found",
			ExpectedEnabled: false,
		},
		{
			Case:            "Sitecore local config",
			ExpectedEnabled: true,
			ExpectedString:  "https://xmcloudcm.local false",
			Template:        "{{ .Environment }} {{ .Cloud }}",
			IsCloud:         false,
		},
		{
			Case:            "Sitecore cloud config",
			ExpectedEnabled: true,
			ExpectedString:  "https://xmc-sitecore<someID>-projectName-environmentName.sitecorecloud.io true",
			Template:        "{{ .Environment }} {{ .Cloud }}",
			IsCloud:         true,
		},
		{
			Case:            "Sitecore cloud config - advanced template",
			ExpectedEnabled: true,
			ExpectedString:  "sitecore<someID> - projectName - environmentName",
			Template:        "{{ if .Cloud }} {{ $splittedHostName := split \".\" .Environment }} {{ $myList := split \"-\" $splittedHostName._0 }} {{ $myList._1 }} - {{ $myList._2 }} - {{ $myList._3 }} {{ end }}", //nolint:lll
			IsCloud:         true,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(poshHome)
		var sitecoreConfigFile string

		if tc.IsCloud {
			content, _ := os.ReadFile("../test/sitecoreUser1.json")
			sitecoreConfigFile = string(content)
		}

		if !tc.IsCloud {
			content, _ := os.ReadFile("../test/sitecoreUser2.json")
			sitecoreConfigFile = string(content)
		}

		var projectDir *platform.FileInfo
		var err error

		if !tc.ExpectedEnabled {
			err = errors.New("no config directory")
			projectDir = nil
		}

		if tc.ExpectedEnabled {
			err = nil
			projectDir = &platform.FileInfo{
				ParentFolder: "SitecoreProjectRoot",
				Path:         filepath.Join("SitecoreProjectRoot", ".sitecore"),
				IsDir:        true,
			}
		}

		env.On("HasParentFilePath", ".sitecore").Return(projectDir, err)
		env.On("HasFilesInDir", filepath.Join("SitecoreProjectRoot", ".sitecore"), "user.json").Return(true)
		env.On("FileContent", filepath.Join("SitecoreProjectRoot", ".sitecore", "user.json")).Return(sitecoreConfigFile)

		sitecore := &Sitecore{
			env: env,
		}

		assert.Equal(t, tc.ExpectedEnabled, sitecore.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, sitecore), tc.Case)
	}
}
