package segments

type Mvn struct {
	Language
}

const mvnToolName = "mvn"

func (m *Mvn) Enabled() bool {
	m.loadSpec()

	mvnw, err := m.env.HasParentFilePath("mvnw", false)
	if err == nil {
		m.tooling[mvnToolName].executable = mvnw.Path
	}

	return m.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (m *Mvn) Activation() Activation {
	m.loadSpec()

	return m.activation()
}

func (m *Mvn) loadSpec() {
	m.extensions = []string{"pom.xml"}
	// Not marked versionCacheable: Enabled() below can swap executable for a
	// project's mvnw wrapper. A wrapper script is typically boilerplate
	// copied verbatim across projects and its own mtime/size can stay
	// unchanged even when .mvn/wrapper/maven-wrapper.properties is bumped to
	// pin a different Maven distribution - so wrapper output depends on that
	// project config, not on the script's own identity. Left off for the
	// non-wrapper "mvn" case too since this is the same *cmd either way.
	m.tooling = map[string]*cmd{
		mvnToolName: {
			executable: mvnToolName,
			args:       []string{versionFlagArg},
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)(?:-(?P<prerelease>[a-z]+-[0-9]+))?))`,
		},
	}
	m.defaultTooling = []string{mvnToolName}
	m.versionURLTemplate = "https://github.com/apache/maven/releases/tag/maven-{{ .Full }}"
}

func (m *Mvn) Template() string {
	return languageTemplate
}
