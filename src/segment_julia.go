package main

type julia struct {
	language
}

func (j *julia) template() string {
	return languageTemplate
}

func (j *julia) init(props Properties, env Environment) {
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
		versionURLTemplate: "https://github.com/JuliaLang/julia/releases/tag/v{{ .Full }})",
	}
}

func (j *julia) enabled() bool {
	return j.language.enabled()
}
