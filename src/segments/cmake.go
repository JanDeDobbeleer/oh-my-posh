package segments

type Cmake struct {
	Language
}

func (c *Cmake) Template() string {
	return languageTemplate
}

const cmakeToolName = "cmake"

func (c *Cmake) Enabled() bool {
	c.extensions = []string{"*.cmake", "CMakeLists.txt"}
	c.tooling = map[string]*cmd{
		cmakeToolName: {
			executable: cmakeToolName,
			args:       []string{versionFlagArg},
			regex:      `cmake version ` + versionRegex,
		},
	}
	c.defaultTooling = []string{cmakeToolName}
	c.versionURLTemplate = "https://cmake.org/cmake/help/v{{ .Major }}.{{ .Minor }}"

	return c.Language.Enabled()
}
