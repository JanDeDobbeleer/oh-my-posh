package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type AzFunc struct {
	language
}

func (az *AzFunc) Template() string {
	return languageTemplate
}

func (az *AzFunc) Init(props properties.Properties, env platform.Environment) {
	az.language = language{
		env:        env,
		props:      props,
		extensions: []string{"host.json", "local.settings.json", "function.json"},
		commands: []*cmd{
			{
				executable: "func",
				args:       []string{"--version"},
				regex:      `(?P<version>[0-9.]+)`,
			},
		},
	}
}

func (az *AzFunc) Enabled() bool {
	return az.language.Enabled()
}
