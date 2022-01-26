package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type php struct {
	language
}

func (p *php) template() string {
	return languageTemplate
}

func (p *php) init(props properties.Properties, env environment.Environment) {
	p.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.php", "composer.json", "composer.lock", ".php-version"},
		commands: []*cmd{
			{
				executable: "php",
				args:       []string{"--version"},
				regex:      `(?:PHP (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "[%[1]s](https://www.php.net/ChangeLog-%[2]s.php#PHP_%[2]s_%[3]s)",
	}
}

func (p *php) enabled() bool {
	return p.language.enabled()
}
