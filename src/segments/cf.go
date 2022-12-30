package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type Cf struct {
	language
}

func (c *Cf) Template() string {
	return languageTemplate
}

func (c *Cf) Init(props properties.Properties, env platform.Environment) {
	c.language = language{
		env:        env,
		props:      props,
		extensions: []string{"manifest.yml", "mta.yaml"},
		commands: []*cmd{
			{
				executable: "cf",
				args:       []string{"version"},
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		displayMode:        props.GetString(DisplayMode, DisplayModeFiles),
		versionURLTemplate: "https://github.com/cloudfoundry/cli/releases/tag/v{{ .Full }}",
	}
}

func (c *Cf) Enabled() bool {
	return c.language.Enabled()
}
