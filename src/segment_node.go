package main

type node struct {
	language *language
}

func (n *node) string() string {
	return n.language.string()
}

func (n *node) init(props *properties, env environmentInfo) {
	n.language = &language{
		env:          env,
		props:        props,
		commands:     []string{"node"},
		versionParam: "--version",
		extensions:   []string{"*.js", "*.ts", "package.json"},
		versionRegex: `(?P<version>[0-9]+.[0-9]+.[0-9]+)`,
	}
}

func (n *node) enabled() bool {
	return n.language.enabled()
}
