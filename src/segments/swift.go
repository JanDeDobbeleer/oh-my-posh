package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Swift struct {
	language
}

func (s *Swift) Template() string {
	return languageTemplate
}

func (s *Swift) Init(props properties.Properties, env platform.Environment) {
	s.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.swift", "*.SWIFT"},
		commands: []*cmd{
			{
				executable: "swift",
				args:       []string{"--version"},
				regex:      `Swift version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+)((.|-)(?P<patch>[0-9]+|dev))?))`,
			},
		},
		versionURLTemplate: "https://github.com/apple/swift/releases/tag/swift-{{ .Full }}-RELEASE",
	}
}

func (s *Swift) Enabled() bool {
	return s.language.Enabled()
}
