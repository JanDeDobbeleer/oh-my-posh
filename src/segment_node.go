package main

type node struct {
	language *language
}

func (n *node) string() string {
	return n.language.string()
}

func (n *node) init(props *properties, env environmentInfo) {
	n.language = &language{
		env:        env,
		props:      props,
		extensions: []string{"*.js", "*.ts", "package.json"},
		commands: []*cmd{
			{
				executable: "node",
				args:       []string{"--version"},
				regex:      `(?:v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "[%[1]s](https://github.com/nodejs/node/blob/master/doc/changelogs/CHANGELOG_V%[2]s.md#%[1]s)",
	}
}

func (n *node) enabled() bool {
	return n.language.enabled()
}
