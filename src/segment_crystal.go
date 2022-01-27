package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type crystal struct {
	language
}

func (c *crystal) template() string {
	return languageTemplate
}

func (c *crystal) init(props properties.Properties, env environment.Environment) {
	c.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.cr", "shard.yml"},
		commands: []*cmd{
			{
				executable: "crystal",
				args:       []string{"--version"},
				regex:      `Crystal (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/crystal-lang/crystal/releases/tag/{{ .Full }}",
	}
}

func (c *crystal) enabled() bool {
	return c.language.enabled()
}
