package main

type angular struct {
	language *language
}

func (a *angular) string() string {
	return a.language.string()
}

func (a *angular) init(props *properties, env environmentInfo) {
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

func (a *angular) enabled() bool {
	return a.language.enabled()
}
