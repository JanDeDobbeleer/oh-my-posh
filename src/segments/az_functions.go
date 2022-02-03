package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type AzFunc struct {
	language
}

func (az *AzFunc) Template() string {
	return languageTemplate
}

func (az *AzFunc) Init(props properties.Properties, env environment.Environment) {
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
