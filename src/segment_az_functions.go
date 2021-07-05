package main

import "oh-my-posh/runtime"

type azfunc struct {
	language *language
}

func (az *azfunc) string() string {
	return az.language.string()
}

func (az *azfunc) init(props *properties, env runtime.Environment) {
	az.language = &language{
		env:        env,
		props:      props,
		extensions: []string{"host.json", "local.settings.json"},
		commands: []*cmd{
			{
				executable: "func",
				args:       []string{"--version"},
				regex:      `(?P<version>.+)`,
			},
		},
	}
}

func (az *azfunc) enabled() bool {
	return az.language.enabled()
}
