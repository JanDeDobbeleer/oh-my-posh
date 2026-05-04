package segments

type Bun struct {
	Language
}

func (b *Bun) Template() string {
	return languageTemplate
}

func (b *Bun) Enabled() bool {
	b.extensions = []string{"bun.lockb", "bun.lock"}
	b.tooling = map[string]*cmd{
		bunToolName: {
			executable: bunToolName,
			args:       []string{versionFlagArg},
			regex:      versionRegexPrefixed,
		},
	}
	b.defaultTooling = []string{bunToolName}
	b.versionURLTemplate = "https://github.com/oven-sh/bun/releases/tag/bun-v{{.Full}}"

	return b.Language.Enabled()
}
