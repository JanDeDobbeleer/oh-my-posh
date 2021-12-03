package main

type golang struct {
	language
}

func (g *golang) string() string {
	return g.language.string()
}

func (g *golang) init(props properties, env environmentInfo) {
	g.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.go", "go.mod"},
		commands: []*cmd{
			{
				executable: "go",
				args:       []string{"version"},
				regex:      `(?:go(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)(.(?P<patch>[0-9]+))?)))`,
			},
		},
		versionURLTemplate: "[%s](https://golang.org/doc/go%s.%s)",
	}
}

func (g *golang) enabled() bool {
	return g.language.enabled()
}
