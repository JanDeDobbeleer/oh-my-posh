package segments

type Perl struct {
	language
}

func (p *Perl) Template() string {
	return languageTemplate
}

func (p *Perl) Enabled() bool {
	perlRegex := `This is perl.*v(?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))(?:\.(?P<patch>[0-9]+))?).* built for .+`
	p.extensions = []string{
		".perl-version",
		"*.pl",
		"*.pm",
		"*.t",
	}
	p.commands = []*cmd{
		{
			executable: "perl",
			args:       []string{"-version"},
			regex:      perlRegex,
		},
	}

	return p.language.Enabled()
}
