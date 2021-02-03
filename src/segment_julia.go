package main

type julia struct {
	language *language
}

func (j *julia) string() string {
	return j.language.string()
}

func (j *julia) init(props *properties, env environmentInfo) {
	j.language = &language{
		env:        env,
		props:      props,
		extensions: []string{"*.jl"},
		commands: []*cmd{
			{
				executable: "julia",
				args:       []string{"--version"},
				regex:      `julia version (?P<version>[0-9]+.[0-9]+.[0-9]+)`,
			},
		},
	}
}

func (j *julia) enabled() bool {
	return j.language.enabled()
}
