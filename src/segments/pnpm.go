package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Pnpm struct {
	language
}

func (n *Pnpm) Enabled() bool {
	return n.language.Enabled()
}

func (n *Pnpm) Template() string {
	return " \U000F02C1 {{.Full}} "
}

func (n *Pnpm) Init(props properties.Properties, env runtime.Environment) {
	n.language = language{
		env:        env,
		props:      props,
		extensions: []string{"package.json", "pnpm-lock.yaml"},
		commands: []*cmd{
			{
				executable: "pnpm",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/pnpm/pnpm/releases/tag/v{{ .Full }}",
	}
}
