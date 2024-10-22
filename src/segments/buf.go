package segments

type Buf struct {
	language
}

func (b *Buf) Template() string {
	return languageTemplate
}

func (b *Buf) Enabled() bool {
	b.extensions = []string{"buf.yaml", "buf.gen.yaml", "buf.work.yaml"}
	b.commands = []*cmd{
		{
			executable: "buf",
			args:       []string{"--version"},
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	b.versionURLTemplate = "https://github.com/bufbuild/buf/releases/tag/v{{.Full}}"

	return b.language.Enabled()
}
