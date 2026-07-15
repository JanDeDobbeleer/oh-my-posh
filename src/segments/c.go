package segments

type C struct {
	Language
}

const cToolName string = "gcc"

func (c *C) Enabled() bool {
	c.extensions = []string{"*.c", "*.cpp", "*.h", "CMakeLists.txt"}
	c.defaultTooling = []string{cToolName}
	c.tooling = map[string]*cmd{
		cToolName: {
			executable: cToolName,
			args:       []string{versionFlagArg},
			regex:      `(?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+))`,
		},
	}
	return c.Language.Enabled()
}

func (c *C) Template() string {
	return "{{ if .Error }}{{ else }} \ue7e5 {{ .Full }}{{ end }}"
}
