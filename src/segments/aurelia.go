package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Aurelia struct {
	Icon string
	language
}

const (
	// Aurelia's icon
	Icon properties.Property = "icon"
)

func (a *Aurelia) Template() string {
	return languageTemplate
}

func (a *Aurelia) Enabled() bool {
	a.extensions = []string{"package.json"}
	a.commands = []*cmd{
		{
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(-(?P<prerelease>[a-z]+).(?P<buildmetadata>[0-9]+))?)))`,
			getVersion: a.getVersion,
		},
	}

	a.versionURLTemplate = "https://github.com/aurelia/aurelia/releases/tag/v{{ .Full }}"

	a.Icon = b.props.GetString(Icon, "\u03b1")

	return a.language.Enabled()
}

func (a *Aurelia) getVersion() (string, error) {
	// tested by nx_test.go
	return getNodePackageVersion(r.language.env, "aurelia")
}
