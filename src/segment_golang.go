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
		version: &version{
			regex:       `(?:go(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			urlTemplate: "[%s](https://golang.org/doc/go%s.%s)",
		},
	}
}

func (g *golang) enabled() bool {
	return g.language.enabled()
}
