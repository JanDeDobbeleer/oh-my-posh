package main

type angularcli struct {
	language *language
}

func (a *angularcli) string() string {
	return a.language.string()
}

func (a *angularcli) init(props *properties, env environmentInfo) {
	a.language = &language{
		env:        env,
		props:      props,
		extensions: []string{"angular.json"},
		commands: []*cmd{
			{
				executable: "ng",
				args:       []string{"--version"},
				regex:      `Angular CLI: (?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "[%s](https://github.com/angular/angular/releases/tag/%s.%s.%s)",
	}
}

func (a *angularcli) enabled() bool {
	return a.language.enabled()
}
