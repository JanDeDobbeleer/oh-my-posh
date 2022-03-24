package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Php struct {
	language
}

func (p *Php) Template() string {
	return languageTemplate
}

func (p *Php) Init(props properties.Properties, env environment.Environment) {
	p.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.php", "composer.json", "composer.lock", ".php-version", "blade.php"},
		commands: []*cmd{
			{
				executable: "php",
				args:       []string{"--version"},
				regex:      `(?:PHP (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "https://www.php.net/ChangeLog-{{ .Major }}.php#PHP_{{ .Major }}_{{ .Minor }}",
	}
}

func (p *Php) Enabled() bool {
	return p.language.Enabled()
}
