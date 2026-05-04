package segments

type Perl struct {
	Language
}

func (p *Perl) Template() string {
	return languageTemplate
}

func (p *Perl) Enabled() bool {
	const perlToolName = "perl"

	perlRegex := `This is perl.*v(?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))(?:\.(?P<patch>[0-9]+))?).* built for .+`
	p.extensions = []string{
		".perl-version",
		"*.pl",
		"*.pm",
		"*.t",
	}
	p.tooling = map[string]*cmd{
		perlToolName: {
			executable: perlToolName,
			args:       []string{versionFlagShortArg},
			regex:      perlRegex,
		},
	}
	p.defaultTooling = []string{perlToolName}

	return p.Language.Enabled()
}
