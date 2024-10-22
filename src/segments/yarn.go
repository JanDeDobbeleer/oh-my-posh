package segments

type Yarn struct {
	language
}

func (n *Yarn) Template() string {
	return " \U000F011B {{.Full}} "
}

func (n *Yarn) Enabled() bool {
	n.extensions = []string{"package.json", "yarn.lock"}
	n.commands = []*cmd{
		{
			executable: "yarn",
			args:       []string{"--version"},
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	n.versionURLTemplate = "https://github.com/yarnpkg/berry/releases/tag/v{{ .Full }}"

	return n.language.Enabled()
}
