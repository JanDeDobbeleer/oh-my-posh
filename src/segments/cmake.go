package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Cmake struct {
	language
}

func (c *Cmake) Template() string {
	return languageTemplate
}

func (c *Cmake) Init(props properties.Properties, env platform.Environment) {
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
