package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Yarn struct {
	language
}

func (n *Yarn) Enabled() bool {
	return n.language.Enabled()
}

func (n *Yarn) Template() string {
	return " \U000F011B {{.Full}} "
}

func (n *Yarn) Init(props properties.Properties, env platform.Environment) {
	n.language = language{
		env:        env,
		props:      props,
		extensions: []string{"package.json", "yarn.lock"},
		commands: []*cmd{
			{
				executable: "yarn",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/yarnpkg/berry/releases/tag/v{{ .Full }}",
	}
}
