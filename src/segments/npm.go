package segments

type Npm struct {
	Language
}

func (n *Npm) Enabled() bool {
	n.extensions = []string{"package.json", "package-lock.json"}
	n.tooling = map[string]*cmd{
		"npm": {
			executable: "npm",
			args:       []string{"--version"},
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	n.defaultTooling = []string{"npm"}
	n.versionURLTemplate = "https://github.com/npm/cli/releases/tag/v{{ .Full }}"

	return n.Language.Enabled()
}

func (n *Npm) Template() string {
	return " \ue71e {{.Full}} "
}
