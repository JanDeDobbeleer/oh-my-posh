package segments

type Nim struct {
	Language
}

func (n *Nim) Template() string {
	return languageTemplate
}

func (n *Nim) Enabled() bool {
	const nimToolName = "nim"

	n.extensions = []string{"*.nim", "*.nims"}

	n.tooling = map[string]*cmd{
		nimToolName: {
			executable: nimToolName,
			args:       []string{versionFlagArg},
			regex:      `Nim Compiler Version (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+))`,
		},
	}
	n.defaultTooling = []string{nimToolName}
	return n.Language.Enabled()
}
