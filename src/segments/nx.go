package segments

import (
	"encoding/json"
	"fmt"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"path/filepath"
)

type Nx struct {
	language
}

func (a *Nx) Template() string {
	return languageTemplate
}

func (a *Nx) Init(props properties.Properties, env platform.Environment) {
	a.language = language{
		env:        env,
		props:      props,
		extensions: []string{"workspace.json", "nx.json"},
		commands: []*cmd{
			{
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
				getVersion: a.getVersion,
			},
		},
		versionURLTemplate: "https://github.com/nrwl/nx/releases/tag/{{.Full}}",
	}
}

func (a *Nx) Enabled() bool {
	return a.language.Enabled()
}

func (a *Nx) getVersion() (string, error) {
	return getNodePackageVersion(a.language.env, "nx")
}

func getNodePackageVersion(env platform.Environment, nodePackage string) (string, error) {
	const fileName string = "package.json"
	folder := filepath.Join(env.Pwd(), "node_modules", nodePackage)
	if !env.HasFilesInDir(folder, fileName) {
		return "", fmt.Errorf("%s not found in %s", fileName, folder)
	}
	content := env.FileContent(filepath.Join(folder, fileName))
	var data ProjectData
	err := json.Unmarshal([]byte(content), &data)
	if err != nil {
		return "", err
	}
	return data.Version, nil
}
