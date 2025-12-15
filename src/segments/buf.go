package segments

type Buf struct {
	Language
}

func (b *Buf) Template() string {
	return languageTemplate
}

func (b *Buf) Enabled() bool {
	b.extensions = []string{"buf.yaml", "buf.gen.yaml", "buf.work.yaml"}
	b.tooling = map[string]*cmd{
		"buf": {
			executable: "buf",
			args:       []string{"--version"},
			regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
		},
	}
	b.defaultTooling = []string{"buf"}
	b.versionURLTemplate = "https://github.com/bufbuild/buf/releases/tag/v{{.Full}}"

	return b.Language.Enabled()
}
