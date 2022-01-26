package main

import "oh-my-posh/environment"

type azfunc struct {
	language
}

func (az *azfunc) template() string {
	return languageTemplate
}

func (az *azfunc) init(props Properties, env environment.Environment) {
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

func (az *azfunc) enabled() bool {
	return az.language.enabled()
}
