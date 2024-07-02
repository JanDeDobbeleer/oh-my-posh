package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type R struct {
	language
}

func (r *R) Template() string {
	return languageTemplate
}

func (r *R) Init(props properties.Properties, env runtime.Environment) {
	rRegex := `version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`
	r.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.R", "*.Rmd", "*.Rsx", "*.Rda", "*.Rd", "*.Rproj", ".Rproj.user"},
		commands: []*cmd{
			{
				executable: "Rscript",
				args:       []string{"--version"},
				regex:      rRegex,
			},
			{
				executable: "R",
				args:       []string{"--version"},
				regex:      rRegex,
			},
			{
				executable: "R.exe",
				args:       []string{"--version"},
				regex:      rRegex,
			},
		},
		versionURLTemplate: "https://www.r-project.org/",
	}
}

func (r *R) Enabled() bool {
	return r.language.Enabled()
}
