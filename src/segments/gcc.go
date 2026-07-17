package segments

type Gcc struct {
	Language
}

const gccToolName string = "gcc"

func (gcc *Gcc) Enabled() bool {
	gcc.extensions = []string{"*.c", "*.cpp", "*.h", "CMakeLists.txt"}
	gcc.defaultTooling = []string{gccToolName}
	gcc.tooling = map[string]*cmd{
		gccToolName: {
			executable: gccToolName,
			args:       []string{versionFlagArg},
			regex:      `(?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+))`,
		},
	}
	return gcc.Language.Enabled()
}

func (gcc *Gcc) Template() string {
	return "{{ if .Error }}{{ else }} \ue7e5 {{ .Full }}{{ end }}"
}
