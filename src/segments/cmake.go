package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Cmake struct {
	language
}

func (c *Cmake) Template() string {
	return languageTemplate
}

func (c *Cmake) Init(props properties.Properties, env environment.Environment) {
	c.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.cmake", "CMakeLists.txt"},
		commands: []*cmd{
			{
				executable: "cmake",
				args:       []string{"--version"},
				regex:      `cmake version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://cmake.org/cmake/help/v{{ .Major }}.{{ .Minor }}",
	}
}

func (c *Cmake) Enabled() bool {
	return c.language.Enabled()
}
