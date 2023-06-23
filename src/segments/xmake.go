package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type XMake struct {
	language
}

func (x *XMake) Template() string {
	return languageTemplate
}

func (x *XMake) Init(props properties.Properties, env platform.Environment) {
	x.language = language{
		env:        env,
		props:      props,
		extensions: []string{"xmake.lua"},
		commands: []*cmd{
			{
				executable: "xmake",
				args:       []string{"--version"},
				regex:      `xmake v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
	}
}

func (x *XMake) Enabled() bool {
	return x.language.Enabled()
}
