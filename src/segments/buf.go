package segments

type Buf struct {
	Language
}

func (b *Buf) Template() string {
	return languageTemplate
}

const bufToolName = "buf"

func (b *Buf) Enabled() bool {
	b.loadSpec()

	return b.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (b *Buf) Activation() Activation {
	b.loadSpec()

	return b.activation()
}

func (b *Buf) loadSpec() {
	b.extensions = []string{"buf.yaml", "buf.gen.yaml", "buf.work.yaml"}
	b.tooling = map[string]*cmd{
		bufToolName: {
			executable:       bufToolName,
			args:             []string{versionFlagArg},
			regex:            versionRegexPrefixed,
			versionCacheable: true,
		},
	}
	b.defaultTooling = []string{bufToolName}
	b.versionURLTemplate = "https://github.com/bufbuild/buf/releases/tag/v{{.Full}}"
}
