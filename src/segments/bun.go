package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Bun struct {
	language
}

func (b *Bun) Template() string {
	return languageTemplate
}

func (b *Bun) Init(props properties.Properties, env platform.Environment) {
	b.language = language{
		env:        env,
		props:      props,
		extensions: []string{"bun.lockb"},
		commands: []*cmd{
			{
				executable: "bun",
				args:       []string{"--version"},
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "https://github.com/oven-sh/bun/releases/tag/bun-v{{.Full}}",
	}
}

func (b *Bun) Enabled() bool {
	return b.language.Enabled()
}
