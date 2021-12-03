package main

type julia struct {
	language
}

func (j *julia) string() string {
	return j.language.string()
}

func (j *julia) init(props properties, env environmentInfo) {
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
		versionURLTemplate: "[%s](https://github.com/JuliaLang/julia/releases/tag/v%s.%s.%s)",
	}
}

func (j *julia) enabled() bool {
	return j.language.enabled()
}
