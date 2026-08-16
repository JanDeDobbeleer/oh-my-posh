package segments

type XMake struct {
	Language
}

func (x *XMake) Template() string {
	return languageTemplate
}

func (x *XMake) Enabled() bool {
	x.loadSpec()

	return x.Language.Enabled()
}

// Activation implements the activation gate; see Language.activation.
func (x *XMake) Activation() Activation {
	x.loadSpec()

	return x.activation()
}

func (x *XMake) loadSpec() {
	const xmakeToolName = "xmake"

	x.extensions = []string{"xmake.lua"}
	x.tooling = map[string]*cmd{
		xmakeToolName: {
			executable: xmakeToolName,
			args:       []string{versionFlagArg},
			regex:      `xmake v(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	x.defaultTooling = []string{xmakeToolName}
}
