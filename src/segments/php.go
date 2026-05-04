package segments

type Php struct {
	Language
}

func (p *Php) Template() string {
	return languageTemplate
}

func (p *Php) Enabled() bool {
	p.extensions = []string{"*.php", "composer.json", "composer.lock", ".php-version", "blade.php"}
	p.tooling = map[string]*cmd{
		phpToolName: {
			executable: phpToolName,
			args:       []string{versionFlagArg},
			regex:      `(?:PHP (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	p.defaultTooling = []string{phpToolName}
	p.versionURLTemplate = "https://www.php.net/ChangeLog-{{ .Major }}.php#PHP_{{ .Major }}_{{ .Minor }}"

	return p.Language.Enabled()
}
