package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type AzFunc struct {
	language
}

func (az *AzFunc) template() string {
	return languageTemplate
}

func (az *AzFunc) init(props properties.Properties, env environment.Environment) {
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

func (az *AzFunc) enabled() bool {
	return az.language.enabled()
}
