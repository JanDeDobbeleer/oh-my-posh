package segments

type Cmake struct {
	Language
}

func (c *Cmake) Template() string {
	return languageTemplate
}

const cmakeToolName = "cmake"

func (c *Cmake) Enabled() bool {
	c.loadSpec()

	return c.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (c *Cmake) Activation() Activation {
	c.loadSpec()

	return c.activation()
}

func (c *Cmake) loadSpec() {
	c.extensions = []string{"*.cmake", "CMakeLists.txt"}
	c.tooling = map[string]*cmd{
		cmakeToolName: {
			executable:       cmakeToolName,
			args:             []string{versionFlagArg},
			regex:            `cmake version ` + versionRegex,
			versionCacheable: true,
		},
	}
	c.defaultTooling = []string{cmakeToolName}
	c.versionURLTemplate = "https://cmake.org/cmake/help/v{{ .Major }}.{{ .Minor }}"
}
