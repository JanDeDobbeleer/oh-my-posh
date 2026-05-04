package segments

type Buf struct {
	Language
}

func (b *Buf) Template() string {
	return languageTemplate
}

const bufToolName = "buf"

func (b *Buf) Enabled() bool {
	b.extensions = []string{"buf.yaml", "buf.gen.yaml", "buf.work.yaml"}
	b.tooling = map[string]*cmd{
		bufToolName: {
			executable: bufToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegexPrefixed,
		},
	}
	b.defaultTooling = []string{bufToolName}
	b.versionURLTemplate = "https://github.com/bufbuild/buf/releases/tag/v{{.Full}}"

	return b.Language.Enabled()
}
