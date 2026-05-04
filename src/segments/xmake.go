package segments

type XMake struct {
	Language
}

func (x *XMake) Template() string {
	return languageTemplate
}

func (x *XMake) Enabled() bool {
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

	return x.Language.Enabled()
}
