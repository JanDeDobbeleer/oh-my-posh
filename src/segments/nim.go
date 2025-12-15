package segments

type Nim struct {
	Language
}

func (n *Nim) Template() string {
	return languageTemplate
}

func (n *Nim) Enabled() bool {
	n.extensions = []string{"*.nim", "*.nims"}

	n.tooling = map[string]*cmd{
		"nim": {
			executable: "nim",
			args:       []string{"--version"},
			regex:      `Nim Compiler Version (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+))`,
		},
	}
	n.defaultTooling = []string{"nim"}
	return n.Language.Enabled()
}
