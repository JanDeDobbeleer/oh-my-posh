package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Vue struct {
	language
}

func (r *Vue) Template() string {
	return languageTemplate
}

func (r *Vue) Init(props properties.Properties, env runtime.Environment) {
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
		versionURLTemplate: "https://github.com/vuejs/core/releases/tag/v{{.Full}}",
	}
}

func (r *Vue) Enabled() bool {
	return r.language.Enabled()
}

func (r *Vue) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(r.language.env, "vue")
}
