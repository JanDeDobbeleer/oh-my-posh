package segments

type Pnpm struct {
	language
}

func (n *Pnpm) Enabled() bool {
	n.extensions = []string{"package.json", "pnpm-lock.yaml"}
	n.commands = []*cmd{
		{
			executable: "pnpm",
			args:       []string{"--version"},
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	n.versionURLTemplate = "https://github.com/pnpm/pnpm/releases/tag/v{{ .Full }}"

	return n.language.Enabled()
}

func (n *Pnpm) Template() string {
	return " \U000F02C1 {{.Full}} "
}
