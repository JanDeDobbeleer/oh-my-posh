package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Nuxt struct {
	language
}

func (r *Nuxt) Template() string {
	return languageTemplate
}

func (r *Nuxt) Init(props properties.Properties, env runtime.Environment) {
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
		versionURLTemplate: "https://github.com/nuxt/nuxt/releases/tag/v{{.Full}}",
	}
}

func (r *Nuxt) Enabled() bool {
	return r.language.Enabled()
}

func (r *Nuxt) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(r.language.env, "nuxt")
}
