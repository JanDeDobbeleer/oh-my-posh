package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Php struct {
	language
}

func (p *Php) template() string {
	return languageTemplate
}

func (p *Php) init(props properties.Properties, env environment.Environment) {
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

func (p *Php) enabled() bool {
	return p.language.enabled()
}
