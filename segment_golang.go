package main

type golang struct {
	language *language
}

func (g *golang) string() string {
	return g.language.string()
}

func (g *golang) init(props *properties, env environmentInfo) {
	g.language = &language{
		env:          env,
		props:        props,
		commands:     []string{"go"},
		versionParam: "version",
		extensions:   []string{"*.go", "go.mod"},
		versionRegex: `go(?P<version>[0-9]+.[0-9]+.[0-9]+)`,
	}
}

func (g *golang) enabled() bool {
	return g.language.enabled()
}
