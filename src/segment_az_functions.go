package main

type azfunc struct {
	language
}

func (az *azfunc) string() string {
	return az.language.string()
}

func (az *azfunc) init(props Properties, env Environment) {
	az.language = language{
		env:        env,
		props:      props,
		extensions: []string{"host.json", "local.settings.json", "function.json"},
		commands: []*cmd{
			{
				executable: "func",
				args:       []string{"--version"},
				regex:      `(?P<version>[0-9.]+)`,
			},
		},
	}
}

func (az *azfunc) enabled() bool {
	return az.language.enabled()
}
