package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type Perl struct {
	language
}

func (p *Perl) Template() string {
	return languageTemplate
}

func (p *Perl) Init(props properties.Properties, env platform.Environment) {
	perlRegex := `This is perl.*v(?P<version>(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))(?:\.(?P<patch>[0-9]+))?).* built for .+`
	p.language = language{
		env:   env,
		props: props,
		extensions: []string{
			".perl-version",
			"*.pl",
			"*.pm",
			"*.t",
		},
		commands: []*cmd{
			{
				executable: "perl",
				args:       []string{"-version"},
				regex:      perlRegex,
			},
		},
	}
}

func (p *Perl) Enabled() bool {
	return p.language.Enabled()
}
