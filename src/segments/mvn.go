package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type Mvn struct {
	language
}

func (m *Mvn) Enabled() bool {
	return m.language.Enabled()
}

func (m *Mvn) Template() string {
	return languageTemplate
}

func (m *Mvn) Init(props properties.Properties, env runtime.Environment) {
	m.language = language{
		env:        env,
		props:      props,
		extensions: []string{"pom.xml"},
		commands: []*cmd{
			{
				executable: "mvn",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(?:-(?P<prerelease>[a-z]+-[0-9]+))?))`,
			},
		},
		versionURLTemplate: "https://github.com/apache/maven/releases/tag/maven-{{ .Full }}",
	}

	mvnw, err := m.language.env.HasParentFilePath("mvnw", false)
	if err == nil {
		m.language.commands[0].executable = mvnw.Path
	}
}
