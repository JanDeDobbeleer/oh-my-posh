package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
	"path/filepath"
)

type Angular struct {
	language
}

func (a *Angular) Template() string {
	return languageTemplate
}

func (a *Angular) Init(props properties.Properties, env platform.Environment) {
	a.language = language{
		env:        env,
		props:      props,
		extensions: []string{"angular.json"},
		commands: []*cmd{
			{
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
				getVersion: a.getVersion,
			},
		},
		versionURLTemplate: "https://github.com/angular/angular/releases/tag/{{.Full}}",
	}
}

func (a *Angular) Enabled() bool {
	return a.language.Enabled()
}

func (a *Angular) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(a.language.env, filepath.Join("@angular", "core"))
}
