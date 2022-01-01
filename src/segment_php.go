package main

type php struct {
	language
}

func (n *php) string() string {
	return n.language.string()
}

func (n *php) init(props Properties, env environmentInfo) {
	n.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.php", "composer.json", "composer.lock", ".php-version"},
		commands: []*cmd{
			{
				executable: "php",
				args:       []string{"--version"},
				regex:      `(?:PHP (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "[%[1]s](https://www.php.net/ChangeLog-%[2]s.php#PHP_%[2]s_%[3]s)",
	}
}

func (n *php) enabled() bool {
	return n.language.enabled()
}
