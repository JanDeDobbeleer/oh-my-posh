package segments

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
)

func TestQuasar(t *testing.T) {
	packageLockFile := `{
		"name": "quasar-project",
		"version": "0.0.1",
		"lockfileVersion": 2,
		"requires": true,
		"dependencies": {
			"@quasar/app-vite": {
				"version": "1.4.3",
				"resolved": "https://registry.npmjs.org/@quasar/app-vite/-/app-vite-1.4.3.tgz",
				"integrity": "sha512-5iMs1sk6fyYTFoRVySwFXWL/PS23UEsdk+YSFejhXnSs5fVDmb2GQMguCHwDl3jPIHjZ7A+XKkb2iWx9pjiPXw==",
				"dev": true
			},
			"vite": {
				"version": "2.9.16",
				"resolved": "https://registry.npmjs.org/vite/-/vite-2.9.16.tgz",
				"integrity": "sha512-X+6q8KPyeuBvTQV8AVSnKDvXoBMnTx8zxh54sOwmmuOdxkjMmEJXH2UEchA+vTMps1xw9vL64uwJOWryULg7nA==",
				"dev": true,
				"requires": {
					"esbuild": "^0.14.27",
					"fsevents": "~2.3.2",
					"postcss": "^8.4.13",
					"resolve": "^1.22.0",
					"rollup": ">=2.59.0 <2.78.0"
				}
			}
		}
	}`

	cases := []struct {
		Case               string
		ExpectedString     string
		Version            string
		HasPackageLockFile bool
		FetchDependencies  bool
	}{
		{Case: "@quasar/cli v2.2.1", ExpectedString: "\uea6a 2.2.1", Version: "@quasar/cli v2.2.1"},
		{
			Case:               "@quasar/cli v2.2.1 with vite",
			Version:            "@quasar/cli v2.2.1",
			HasPackageLockFile: true,
			FetchDependencies:  true,
			ExpectedString:     "\uea6a 2.2.1 \ueb29 2.9.16",
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "quasar",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "quasar.config",
		}
		env, props := getMockedLanguageEnv(params)

		env.On("HasFilesInDir", "/usr/home/project", "package-lock.json").Return(tc.HasPackageLockFile)
		fileInfo := &platform.FileInfo{ParentFolder: "/usr/home/project", IsDir: true}
		env.On("HasParentFilePath", "quasar.config").Return(fileInfo, nil)
		env.On("FileContent", filepath.Join(fileInfo.ParentFolder, "package-lock.json")).Return(packageLockFile)

		props[FetchDependencies] = tc.FetchDependencies

		quasar := &Quasar{}
		quasar.Init(props, env)

		assert.True(t, quasar.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, quasar.Template(), quasar), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
