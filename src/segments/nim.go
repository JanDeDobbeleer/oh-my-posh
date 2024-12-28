package segments

type Nim struct {
	language
}

func (n *Nim) Template() string {
	return languageTemplate
}

func (n *Nim) Enabled() bool {
	n.extensions = []string{"*.nim", "*.nims"}

	n.commands = []*cmd{
		{
			executable: "nim",
			args:       []string{"--version"},
			regex:      `Nim Compiler Version (?P<version>(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+))`,
		},
	}
	return n.language.Enabled()
}
