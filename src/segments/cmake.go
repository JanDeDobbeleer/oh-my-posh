package segments

type Cmake struct {
	Language
}

func (c *Cmake) Template() string {
	return languageTemplate
}

func (c *Cmake) Enabled() bool {
	c.extensions = []string{"*.cmake", "CMakeLists.txt"}
	c.tooling = map[string]*cmd{
		"cmake": {
			executable: "cmake",
			args:       []string{"--version"},
			regex:      `cmake version (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	c.defaultTooling = []string{"cmake"}
	c.versionURLTemplate = "https://cmake.org/cmake/help/v{{ .Major }}.{{ .Minor }}"

	return c.Language.Enabled()
}
