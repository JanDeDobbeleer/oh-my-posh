package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type React struct {
	language
}

func (r *React) Template() string {
	return languageTemplate
}

func (r *React) Init(props properties.Properties, env platform.Environment) {
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"package.json"},
		commands: []*cmd{
			{
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
				getVersion: r.getVersion,
			},
		},
		versionURLTemplate: "https://github.com/facebook/react/releases/tag/v{{.Full}}",
	}
}

func (r *React) Enabled() bool {
	return r.language.Enabled()
}

func (r *React) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(r.language.env, "react")
}
