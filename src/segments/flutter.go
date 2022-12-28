package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/platform"
	"github.com/jandedobbeleer/oh-my-posh/properties"
)

type Flutter struct {
	language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

func (f *Flutter) Init(props properties.Properties, env platform.Environment) {
	f.language = language{
		env:        env,
		props:      props,
		extensions: dartExtensions,
		folders:    dartFolders,
		commands: []*cmd{
			{
				executable: "flutter",
				args:       []string{"--version"},
				regex:      `Flutter (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/flutter/flutter/releases/tag/{{ .Major }}.{{ .Minor }}.{{ .Patch }}",
	}
}

func (f *Flutter) Enabled() bool {
	return f.language.Enabled()
}
