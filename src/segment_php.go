package main

type php struct {
	language
}

func (p *php) string() string {
	segmentTemplate := p.language.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) == 0 {
		return p.language.string()
	}
	return p.language.renderTemplate(segmentTemplate, p)
}

func (p *php) init(props Properties, env Environment) {
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
