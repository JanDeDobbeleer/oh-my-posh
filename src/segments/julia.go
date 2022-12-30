package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type Julia struct {
	language
}

func (j *Julia) Template() string {
	return languageTemplate
}

func (j *Julia) Init(props properties.Properties, env platform.Environment) {
	j.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.jl"},
		commands: []*cmd{
			{
				executable: "julia",
				args:       []string{"--version"},
				regex:      `julia version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/JuliaLang/julia/releases/tag/v{{ .Full }}",
	}
}

func (j *Julia) Enabled() bool {
	return j.language.Enabled()
}
