package segments

type Yarn struct {
	Language
}

func (n *Yarn) Template() string {
	return " \ue6a7 {{.Full}} "
}

func (n *Yarn) Enabled() bool {
	n.extensions = []string{"package.json", "yarn.lock"}
	n.tooling = map[string]*cmd{
		"yarn": {
			executable: "yarn",
			args:       []string{"--version"},
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	n.defaultTooling = []string{"yarn"}
	n.versionURLTemplate = "https://github.com/yarnpkg/berry/releases/tag/v{{ .Full }}"

	return n.Language.Enabled()
}
