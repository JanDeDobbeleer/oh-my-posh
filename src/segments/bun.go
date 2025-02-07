package segments

type Bun struct {
	language
}

func (b *Bun) Template() string {
	return languageTemplate
}

func (b *Bun) Enabled() bool {
	b.extensions = []string{"bun.lockb", "bun.lock"}
	b.commands = []*cmd{
		{
			executable: "bun",
			args:       []string{"--version"},
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	b.versionURLTemplate = "https://github.com/oven-sh/bun/releases/tag/bun-v{{.Full}}"

	return b.language.Enabled()
}
