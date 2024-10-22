package segments

type Mvn struct {
	language
}

func (m *Mvn) Enabled() bool {
	m.extensions = []string{"pom.xml"}
	m.commands = []*cmd{
		{
			executable: "mvn",
			args:       []string{"--version"},
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(?:-(?P<prerelease>[a-z]+-[0-9]+))?))`,
		},
	}
	m.versionURLTemplate = "https://github.com/apache/maven/releases/tag/maven-{{ .Full }}"

	mvnw, err := m.env.HasParentFilePath("mvnw", false)
	if err == nil {
		m.commands[0].executable = mvnw.Path
	}

	return m.language.Enabled()
}

func (m *Mvn) Template() string {
	return languageTemplate
}
