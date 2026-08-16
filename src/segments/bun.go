package segments

type Bun struct {
	Language
}

func (b *Bun) Template() string {
	return languageTemplate
}

func (b *Bun) Enabled() bool {
	b.loadSpec()

	return b.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (b *Bun) Activation() Activation {
	b.loadSpec()

	return b.activation()
}

func (b *Bun) loadSpec() {
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
}
