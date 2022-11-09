package segments

import (
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

type Kotlin struct {
	language
}

func (k *Kotlin) Template() string {
	return languageTemplate
}

func (k *Kotlin) Init(props properties.Properties, env platform.Environment) {
	k.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.kt", "*.kts", "*.ktm"},
		commands: []*cmd{
			{
				executable: "kotlin",
				args:       []string{"-version"},
				regex:      `Kotlin version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/JetBrains/kotlin/releases/tag/v{{ .Full }}",
	}
}

func (k *Kotlin) Enabled() bool {
	return k.language.Enabled()
}
